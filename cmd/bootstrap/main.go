package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

var (
	reBootstrap    = regexp.MustCompile(`^bootstrap\s`)
	reEmptyLine    = regexp.MustCompile(`^\s*$`)
	reRoot         = regexp.MustCompile(`^ROOT=\S+`)
	reShelley      = regexp.MustCompile(`sed\s.*shelley/genesis.spec.json`)
	reTestnetMagic = regexp.MustCompile(`^NETWORK_MAGIC=\S+`)
)

type Param struct {
	Path  []string
	Raw   string
	Value interface{}
}

type Config struct {
	Del     []Param
	Numbers []Param
	Strings []Param
	Set     []Param
}

func (c Config) Merge(that Config) Config {
	return Config{
		Del:     append(c.Del, that.Del...),
		Numbers: append(c.Numbers, that.Numbers...),
		Strings: append(c.Strings, that.Strings...),
		Set:     append(c.Set, that.Set...),
	}
}

type Options struct {
	Del     cli.StringSlice
	Numbers cli.StringSlice
	Set     cli.StringSlice
	Strings cli.StringSlice
}

func (o Options) Parse() (Config, error) {
	del, err := parseAllParam(o.Del.Value())
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse options: %w", err)
	}

	num, err := parseAllParam(o.Numbers.Value())
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse options: %w", err)
	}

	set, err := parseAllParam(o.Set.Value())
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse options: %w", err)
	}

	str, err := parseAllParam(o.Strings.Value())
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse options: %w", err)
	}

	return Config{
		Del:     del,
		Numbers: num,
		Strings: str,
		Set:     set,
	}, nil
}

var opts struct {
	DryRun       bool
	Filename     string
	Root         string
	TestnetMagic string
	WriteInPlace bool
	Alonzo       Options
	Common       Options
	Path         struct {
		CardanoNode string
	}
}

var flags = []cli.Flag{
	&cli.StringSliceFlag{
		Name:        "del",
		Usage:       "delete path",
		Value:       &cli.StringSlice{},
		Destination: &opts.Common.Del,
	},
	&cli.StringFlag{
		Name:        "f,file",
		Usage:       "path to cardano-node:scripts/byron-to-alonzo/mkfiles.sh",
		Value:       os.ExpandEnv("${HOME}/src/cardano-node/scripts/byron-to-alonzo/mkfiles.sh"),
		Destination: &opts.Filename,
	},
	&cli.StringSliceFlag{
		Name:        "n,num",
		Usage:       "replace numeric value in shelley genesis",
		Value:       &cli.StringSlice{},
		Destination: &opts.Common.Numbers,
	},
	&cli.StringSliceFlag{
		Name:        "s,string",
		Usage:       "replace string value in shelley genesis",
		Value:       &cli.StringSlice{},
		Destination: &opts.Common.Strings,
	},
	&cli.StringSliceFlag{
		Name:        "set",
		Usage:       "set path; requires full path (no partials)",
		Value:       &cli.StringSlice{},
		Destination: &opts.Common.Set,
	},
}

func main() {
	app := cli.NewApp()
	app.Usage = "generate configuration and keys for a private alonzo testnet"
	app.Commands = []*cli.Command{
		{
			Name:   "mkfiles",
			Usage:  "execute mkfiles",
			Action: mkfilesAction,
			Flags: append(flags,
				&cli.BoolFlag{
					Name:        "dry",
					Usage:       "dry run; do not execute script",
					Destination: &opts.DryRun,
				},
				&cli.StringFlag{
					Name:        "m,magic",
					Usage:       "assign testnet-magic",
					Value:       "42",
					Destination: &opts.TestnetMagic,
				},
				&cli.StringFlag{
					Name:        "r,root",
					Usage:       "root directory for alonzo data files",
					Value:       os.ExpandEnv("${HOME}/alonzo-testnet"),
					Destination: &opts.Root,
				},
				&cli.StringFlag{
					Name:        "src",
					Usage:       "path to cardano-node source",
					Value:       os.ExpandEnv("${HOME}/src/cardano-node"),
					Destination: &opts.Path.CardanoNode,
				},
				&cli.StringSliceFlag{
					Name:        "alonzo-del",
					Usage:       "delete path",
					Value:       &cli.StringSlice{},
					Destination: &opts.Alonzo.Del,
				},
				&cli.StringSliceFlag{
					Name:        "alonzo-set",
					Usage:       "set path; requires full path (no partials)",
					Value:       &cli.StringSlice{},
					Destination: &opts.Alonzo.Set,
				},
			),
		},
		{
			Name:   "replace",
			Usage:  "replace parameters in json file",
			Action: replaceAction,
			Flags: append(flags,
				&cli.BoolFlag{
					Name:        "i",
					Usage:       "overwrite existing file",
					Destination: &opts.WriteInPlace,
				},
			),
		},
	}
	app.Flags = flags
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

func mkfilesAction(_ *cli.Context) error {
	config, err := opts.Common.Parse()
	if err != nil {
		return err
	}

	alonzo, err := opts.Alonzo.Parse()
	if err != nil {
		return err
	}

	s, err := rewriteScript(opts.Filename, opts.Root, opts.TestnetMagic, config)
	if err != nil {
		return err
	}

	filename := opts.Filename + ".bootstrap"
	if err := ioutil.WriteFile(filename, []byte(s), 0644); err != nil {
		return fmt.Errorf("failed to rewrite %v: %w", opts.Filename, err)
	}

	if opts.DryRun {
		fmt.Println(s)
		return nil
	}

	cmd := exec.Command("/bin/bash", filename, "alonzo")
	cmd.Dir = opts.Path.CardanoNode
	cmd.Stdout = os.Stdout
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute script, %v: %w", filename, err)
	}

	// post mkfiles commands
	alonzoFilename := filepath.Join(opts.Root, "shelley/genesis.alonzo.json")
	if err := replace(config.Merge(alonzo), alonzoFilename, true); err != nil {
		return fmt.Errorf("failed to update alonzo configuration: %w", err)
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("To start the alonzo testnet:")
	fmt.Println()
	fmt.Println(`    (nohup "${HOME}/alonzo-testnet/run/all.sh" 2>&1) > /dev/null &`)
	fmt.Println()
	fmt.Println()

	return nil
}

func rewriteScript(filename, root, testnetMagic string, config Config) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("mkfiles failed: %w", err)
	}
	data = bytes.Trim(data, "\n")
	data = append(data, '\n')

	buf := bufio.NewReader(bytes.NewReader(data))
	w := bytes.NewBuffer(nil)
	shelley := 0
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("mkfiles failed: %w", err)
		}

		switch {
		case root != "" && reRoot.MatchString(line):
			fmt.Fprintf(w, "ROOT=\"%v\"\n", root)
			continue
		case reTestnetMagic.MatchString(line):
			fmt.Fprintf(w, "NETWORK_MAGIC=\"%v\"\n", testnetMagic)
			continue
		case reBootstrap.MatchString(line):
			continue // strip out prior bootstrap commands
		case shelley == 0 && reShelley.MatchString(line):
			shelley = 1
		case shelley == 1 && reEmptyLine.MatchString(line):
			shelley = 2
			fmt.Fprintln(w)
			fmt.Fprintln(w, makeCLI(config, "shelley/genesis.spec.json"))
			fmt.Fprintln(w)
		}

		fmt.Fprintln(w, strings.TrimRight(line, "\n"))
	}

	return w.String(), nil
}

func makeCLI(config Config, filename string) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(os.ExpandEnv("${HOME}/bin/bootstrap replace -i"))
	for _, p := range config.Del {
		buf.WriteString(` --del "`)
		buf.WriteString(p.Raw)
		buf.WriteString(`"`)
	}
	for _, p := range config.Numbers {
		buf.WriteString(` -n "`)
		buf.WriteString(p.Raw)
		buf.WriteString(`"`)
	}
	for _, p := range config.Set {
		buf.WriteString(` --set "`)
		buf.WriteString(p.Raw)
		buf.WriteString(`"`)
	}
	for _, p := range config.Strings {
		buf.WriteString(` -s "`)
		buf.WriteString(p.Raw)
		buf.WriteString(`"`)
	}
	buf.WriteString(" -f ")
	buf.WriteString(filename)
	return buf.String()
}

type replaceFunc func(path []string, v interface{}) (interface{}, bool)

func delFunc(wants []string) replaceFunc {
	return func(path []string, v interface{}) (interface{}, bool) {
		if hasSuffix(path, wants) {
			return nil, true
		}
		return v, false
	}
}

func hasSuffix(path []string, wants []string) bool {
	p := len(path)
	w := len(wants)

	if p < w {
		return false
	}

	for i, want := range wants {
		if path[p-w+i] != want {
			return false
		}
	}

	return true
}

func replaceFloatFunc(wants []string, f float64) replaceFunc {
	return func(path []string, v interface{}) (interface{}, bool) {
		if hasSuffix(path, wants) {
			return f, true
		}
		return v, false
	}
}

func replaceStringFunc(wants []string, s string) replaceFunc {
	return func(path []string, v interface{}) (interface{}, bool) {
		if hasSuffix(path, wants) {
			return s, true
		}
		return v, false
	}
}

func replaceAny(path []string, record interface{}, replace ...replaceFunc) interface{} {
	switch v := record.(type) {
	case map[string]interface{}:
		return replaceMap(path, v, replace...)
	case []interface{}:
		return replaceSlice(path, v, replace...)
	default:
		return record
	}
}

func replaceMap(path []string, record map[string]interface{}, replace ...replaceFunc) interface{} {
	m := map[string]interface{}{}

loop:
	for k, v := range record {
		sub := append(path, k)
		for _, fn := range replace {
			if value, ok := fn(sub, v); ok {
				if value != nil {
					m[k] = value
				}
				continue loop
			}
		}

		m[k] = replaceAny(sub, v, replace...)
	}
	return m
}

func replaceSlice(path []string, records []interface{}, replace ...replaceFunc) interface{} {
	var rr []interface{}
	for _, r := range records {
		rr = append(rr, replaceAny(path, r, replace...))
	}
	return rr
}

func addToMap(parent interface{}, value interface{}, path ...string) interface{} {
	for len(path) > 0 {
		m, ok := parent.(map[string]interface{})
		if !ok {
			return parent
		}

		key := path[0]
		if len(path) == 1 {
			m[key] = value
			return m
		}

		child, ok := m[key]
		if !ok {
			child = map[string]interface{}{}
		}

		m[key] = addToMap(child, value, path[1:]...)
		break
	}

	return parent
}

func parseAllParam(ss []string) ([]Param, error) {
	var pp []Param
	for _, s := range ss {
		p, err := parseParam(s)
		if err != nil {
			return nil, err
		}
		pp = append(pp, p)
	}
	return pp, nil
}

func parseParam(s string) (Param, error) {
	var (
		parts = strings.SplitN(s, "=", 2)
		path  = strings.Split(parts[0], ".")
		value interface{}
	)

	if len(parts) == 2 {
		raw := parts[1]
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			value = v
		} else {
			value = raw
		}
	}

	return Param{
		Path:  path,
		Raw:   s,
		Value: value,
	}, nil
}

func replaceAction(_ *cli.Context) error {
	config, err := opts.Common.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse file, %v: %w", opts.Filename, err)
	}

	return replace(config, opts.Filename, opts.WriteInPlace)
}

func replace(config Config, filename string, writeInPlace bool) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file, %v: %w", filename, err)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to parse file, %v: %w", filename, err)
	}

	v = replaceAny(nil, v, toReplaceFunc(config)...)
	v = setAll(v, config.Set)

	if writeInPlace {
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal file, %v: %w", filename, err)
		}

		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("failed to write file, %v: %w", filename, err)
		}

		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func setAll(v interface{}, pp []Param) interface{} {
	m, ok := v.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("setAll received %T: expected map[string]interface{}", v))
	}

	for _, p := range pp {
		m = addToMap(m, p.Value, p.Path...).(map[string]interface{})
	}

	return m
}

func toReplaceFunc(config Config) (fns []replaceFunc) {
	for _, p := range config.Del {
		fns = append(fns, delFunc(p.Path))
	}
	for _, pp := range [][]Param{config.Numbers, config.Set, config.Strings} {
		for _, p := range pp {
			switch v := p.Value.(type) {
			case float64:
				fns = append(fns, replaceFloatFunc(p.Path, v))
			case string:
				fns = append(fns, replaceStringFunc(p.Path, v))
			}
		}
	}
	return fns
}
