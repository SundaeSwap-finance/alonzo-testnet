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

var opts struct {
	DryRun       bool
	Filename     string
	Numbers      cli.StringSlice
	Root         string
	Strings      cli.StringSlice
	TestnetMagic string
	WriteInPlace bool
	Path         struct {
		CardanoNode string
	}
}

var flags = []cli.Flag{
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
		Destination: &opts.Numbers,
	},
	&cli.StringSliceFlag{
		Name:        "s,string",
		Usage:       "replace string value in shelley genesis",
		Value:       &cli.StringSlice{},
		Destination: &opts.Strings,
	},
}

func main() {
	app := cli.NewApp()
	app.Usage = "generate configuration and keys for a private alonzo testnet"
	app.Commands = []*cli.Command{
		{
			Name:   "mkfiles",
			Usage:  "execute mkfiles",
			Action: mkfiles,
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
			),
		},
		{
			Name:   "replace",
			Usage:  "replace parameters in json file",
			Action: replace,
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

func mkfiles(_ *cli.Context) error {
	s, err := rewriteScript(opts.Filename, opts.Root, opts.TestnetMagic, opts.Numbers.Value(), opts.Strings.Value())
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

	cmd := exec.Command("/bin/sh", filename, "alonzo")
	cmd.Dir = opts.Path.CardanoNode
	cmd.Stdout = os.Stdout
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute script, %v: %w", filename, err)
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

func rewriteScript(filename, root, testnetMagic string, nn, ss []string) (string, error) {
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
			fmt.Fprintln(w, makeCLI(nn, ss))
			fmt.Fprintln(w)
		}

		fmt.Fprintln(w, strings.TrimRight(line, "\n"))
	}

	return w.String(), nil
}

func makeCLI(nn, ss []string) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(os.ExpandEnv("${HOME}/bin/bootstrap replace -i"))
	for _, n := range nn {
		buf.WriteString(` -n "`)
		buf.WriteString(n)
		buf.WriteString(`"`)
	}
	for _, s := range ss {
		buf.WriteString(` -s "`)
		buf.WriteString(s)
		buf.WriteString(`"`)
	}
	buf.WriteString(" -f shelley/genesis.spec.json")
	return buf.String()
}

func replaceAny(record interface{}, replace map[string]interface{}) interface{} {
	switch v := record.(type) {
	case map[string]interface{}:
		return replaceMap(v, replace)
	case []interface{}:
		return replaceSlice(v, replace)
	default:
		return record
	}
}

func replaceMap(record map[string]interface{}, replace map[string]interface{}) interface{} {
	m := map[string]interface{}{}
	for k, v := range record {
		if s, ok := replace[k]; ok {
			m[k] = s
			continue
		}

		m[k] = replaceAny(v, replace)
	}
	return m
}

func replace(_ *cli.Context) error {
	data, err := ioutil.ReadFile(opts.Filename)
	if err != nil {
		return fmt.Errorf("failed to read file, %v: %w", opts.Filename, err)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to parse file, %v: %w", opts.Filename, err)
	}

	v = replaceAny(v, toMap(opts.Numbers.Value(), opts.Strings.Value()))

	if opts.WriteInPlace {
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal file, %v: %w", opts.Filename, err)
		}

		if err := ioutil.WriteFile(opts.Filename, data, 0644); err != nil {
			return fmt.Errorf("failed to write file, %v: %w", opts.Filename, err)
		}

		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func replaceSlice(records []interface{}, replace map[string]interface{}) interface{} {
	var rr []interface{}
	for _, r := range records {
		rr = append(rr, replaceAny(r, replace))
	}
	return rr
}

func toMap(nn, ss []string) map[string]interface{} {
	m := map[string]interface{}{}
	for _, s := range nn {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			k, str := parts[0], parts[1]
			if v, err := strconv.ParseFloat(str, 64); err == nil {
				m[k] = v
			}
		}
	}
	for _, s := range ss {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			k, v := parts[0], parts[1]
			m[k] = v
		}
	}
	return m
}
