package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/TylerBrock/colorjson"
	"github.com/fatih/color"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	"golang.org/x/sys/execabs"
)

// CheckErr prints the msg with the prefix 'Error:' and exits with error code 1. If the msg is nil, it does nothing.
func CheckErr(e error, cmd *cobra.Command) {
	if e != nil {
		fmt.Println()
		fmt.Println()
		_, err := color.New(color.FgRed).Fprintln(os.Stderr, " (╯°□°)╯︵ ɹoɹɹƎ \n\n\n "+e.Error())
		if err != nil {
			color.Red(err.Error())
		}

		fmt.Println()
		fmt.Println()

		if cmd != nil {
			_ = cmd.Help()
		}

		os.Exit(1)
	}
}

func PrintAndReturnError(text string) error {
	color.Red(text)
	return fmt.Errorf(text)
}

func ChunkString(s string, chunkSize int) []string {
	chunks := []string{""}
	words := strings.Fields(s)

	count := 0
	for _, w := range words {
		if len(chunks[count])+len(w) <= chunkSize {
			chunks[count] += " " + w
			continue
		}

		chunks = append(chunks, w)
		count++
	}

	for i, c := range chunks {
		chunks[i] = strings.TrimSpace(c)
	}

	return chunks
}

func Ellipsis(s string, max int) string {
	if max > len(s) {
		return s
	}
	return s[:max] + "..."
}

func DuplicateStrings(arr []string) []string {
	visited := make(map[string]bool, 0)
	var duplicates []string
	for i := 0; i < len(arr); i++ {
		if visited[arr[i]] {
			duplicates = append(duplicates, arr[i])
		} else {
			visited[arr[i]] = true
		}
	}
	return duplicates
}

// DifferenceStrings returns the elements that are in A, but not in B
func DifferenceStrings(a, b []string) []string {
	mb := make(map[string]bool, len(b))
	for _, x := range b {
		mb[x] = true
	}
	var diff []string
	for _, x := range a {
		if !mb[x] {
			diff = append(diff, x)
		}
	}

	return diff
}

// Unique returns the unique element that are in a slice of strings
func Unique(strings []string) []string {
	uniqueMap := map[string]bool{}
	var uniques []string
	for _, s := range strings {
		if !uniqueMap[s] {
			uniques = append(uniques, s)
		}

		uniqueMap[s] = true
	}

	return uniques
}

func Percentage(partialValue, totalValue int) int {
	return (100 * partialValue) / totalValue
}

func GetRealSizeOf(v interface{}) (int, error) {
	b := new(bytes.Buffer)
	if err := gob.NewEncoder(b).Encode(v); err != nil {
		return 0, err
	}
	return b.Len(), nil
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func CreateRangeOfInt(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func ChunkSliceOfInt(slice []int, chunkSize int) [][]int {
	var chunks [][]int
	for {
		if len(slice) == 0 {
			break
		}

		// necessary check to avoid slicing beyond
		// slice capacity
		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}

	return chunks
}

func StringAfter(s, substring string) string {
	split := strings.Split(s, substring)
	if len(split) >= 2 {
		return strings.Join(split[1:], "")
	}

	return s
}

func FormatCommas(num int) string {
	str := fmt.Sprintf("%d", num)
	re := regexp.MustCompile(`(\d+)(\d{3})`)
	for n := ""; n != str; {
		n = str
		str = re.ReplaceAllString(str, "$1,$2")
	}
	return str
}

const escape = "\x1b"

func format(attr color.Attribute) string {
	return fmt.Sprintf("%s[%dm", escape, attr)
}

func PrintYAML(s string) error {
	tokens := lexer.Tokenize(s)
	var p printer.Printer
	p.LineNumber = true
	p.LineNumberFormat = func(num int) string {
		fn := color.New(color.Bold, color.FgBlue).SprintFunc()
		return fn(fmt.Sprintf("%2d | ", num))
	}
	p.Bool = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiBlue),
			Suffix: format(color.Reset),
		}
	}
	p.Number = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgYellow),
			Suffix: format(color.Reset),
		}
	}
	p.MapKey = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiMagenta),
			Suffix: format(color.Reset),
		}
	}
	p.Anchor = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiYellow),
			Suffix: format(color.Reset),
		}
	}
	p.Alias = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiYellow),
			Suffix: format(color.Reset),
		}
	}
	p.String = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgWhite),
			Suffix: format(color.Reset),
		}
	}
	writer := colorable.NewColorableStdout()
	_, err := writer.Write([]byte(p.PrintTokens(tokens) + "\n"))
	return err
}

func IsValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func PrintJSON(withColor bool, body []byte) error {
	if !withColor {
		var out bytes.Buffer
		err := json.Indent(&out, body, "", "\t")
		if err != nil {
			fmt.Println(string(body))
			return nil
		}
		fmt.Println(out.String())
		return nil
	}

	// Create an interesting JSON object to marshal in a pretty format
	var obj interface{}
	err := json.Unmarshal(body, &obj)
	if err != nil {
		return err
	}

	// custom nice colors, the same as `gojq`
	f := colorjson.NewFormatter()
	f.Indent = 2
	b := color.Color{}
	b.Add(color.FgHiBlue)
	b.Set()
	f.KeyColor = &b
	s, _ := f.Marshal(obj)
	fmt.Println(string(s))

	return nil
}

func PrintJSONString(withColor bool, body string) error {
	return PrintJSON(withColor, []byte(body))
}

func ParseJSON(bytes []byte) interface{} {
	var obj interface{}
	err := json.Unmarshal(bytes, &obj)
	if err != nil {
		color.Red("Error the parsing JSON: %s \n", err)
		os.Exit(1)
	}
	return obj
}

func ParseJSONIntoMap(bytes []byte) map[string]interface{} {
	var obj map[string]interface{}
	err := json.Unmarshal(bytes, &obj)
	if err != nil {
		color.Red("Error the parsing JSON: %s \n", err)
		os.Exit(1)
	}
	return obj
}

func ParseInterfaceIntoJSON(yourThing interface{}) []byte {
	obj, err := json.Marshal(yourThing)
	if err != nil {
		color.Red("Error the parsing JSON: %s \n", err)
		os.Exit(1)
	}
	return obj
}

func ParseJSONIntoArray(bytes []byte) []interface{} {
	var obj []interface{}
	err := json.Unmarshal(bytes, &obj)
	if err != nil {
		color.Red("Error the parsing JSON: %s \n", err)
		os.Exit(1)
	}
	return obj
}

func ExecuteCommandAndGetOutput(command string, flags ...string) (string, error) {
	cmd := exec.Command(command, flags...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Start()
	if err != nil {
		return outb.String() + errb.String(), err
	}

	err = cmd.Wait()
	if err != nil {
		return outb.String() + errb.String(), err
	}

	// fmt.Println("out:", outb.String(), "err:", errb.String())

	return outb.String() + errb.String(), nil
}

func Base64DecodeSegment(seg string) ([]byte, error) {
	b, err := base64.URLEncoding.DecodeString(seg)
	if err == nil {
		return b, nil
	}

	// This solved issues in another project when decoding JWT
	// So maybe try it
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}

func OpenURL(url string) {
	_, err := execabs.LookPath("open")
	if err == nil {
		_, _ = ExecuteCommandAndGetOutput("open", url)
	} else {
		_, err = execabs.LookPath("xdg-open")
		if err == nil {
			_, _ = ExecuteCommandAndGetOutput("xdg-open", url)
		}
	}
}

func IsOnPath(tool string) string {
	path, err := execabs.LookPath(tool)
	if err != nil {
		return ""
	}

	if strings.TrimSpace(path) != "" {
		return path
	}

	return ""
}
