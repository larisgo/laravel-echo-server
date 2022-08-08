// Source: https://github.com/ukautz/clif/blob/v1/input.go
package std

import (
	"bufio"
	"fmt"
	"github.com/gookit/color"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Input is an interface for input helping. It provides shorthand methods for
// often used CLI interactions.
type Input interface {

	// Ask prints question to user and then reads user input and returns as soon
	// as it's non empty or queries again until it is
	Ask(question string, check func(string) error) string

	// AskRegex prints question to user and then reads user input, compares it
	// against regex and return if matching or queries again until it does
	AskRegex(question string, rx *regexp.Regexp) string

	// Choose renders choices for user and returns what was choosen
	Choose(question string, choices map[string]string) string

	// Confirm prints question to user until she replies with "yes", "y", "no" or "n"
	Confirm(question string) bool
}

// DefaultInput is the default used input implementation
type DefaultInput struct {
	in  io.Reader
	out Output
}

// NewDefaultInput constructs a new default input implementation on given
// io reader (if nil, fall back to `os.Stdin`). Requires Output for issuing
// questions to user.
func NewDefaultInput(in io.Reader, out Output) *DefaultInput {
	if in == nil {
		in = os.Stdin
	}
	return &DefaultInput{in, out}
}

var (
	RenderAskQuestion = func(question string) string {
		return color.Question.Sprint(strings.TrimRight(question, " "))
	}
	RenderInputRequiredError = fmt.Errorf("Input required")
)

func (i *DefaultInput) Ask(question string, check func(string) error, _default ...string) string {
	_default = append(_default, "")
	if check == nil {
		check = func(in string) error {
			if len(in) > 0 {
				return nil
			} else {
				return RenderInputRequiredError
			}
		}
	}
	reader := bufio.NewReader(i.in)
	_d := ""
	if len(_default[0]) > 0 {
		_d = color.Primary.Sprint(fmt.Sprintf(" (%s) ", _default[0]))
	}
	for {
		i.out.Printf(RenderAskQuestion(question) + _d)
		line, _, err := reader.ReadLine()
		for err == io.EOF {
			<-time.After(time.Millisecond)
			line, _, err = reader.ReadLine()
		}
		if err != nil {
			i.out.Printf("%s\n\n", color.Warn.Sprint(err))
		} else {
			s := string(line)
			if len(s) == 0 {
				s = _default[0]
			}
			if err = check(s); err != nil {
				i.out.Printf("%s\n\n", color.Warn.Sprint(err))
			} else {
				return s
			}
		}
	}
}

// RenderChooseQuestion is the method used by default input `Choose()` method to
// to render the question (displayed before listing the choices) into a string.
// Can be overwritten at users discretion.
var RenderChooseQuestion = func(question string) string {
	return question + "\n"
}

// RenderChooseOption is the method used by default input `Choose()` method to
// to render a singular choice into a string. Can be overwritten at users discretion.
var RenderChooseOption = func(key, value string, size int) string {
	return fmt.Sprintf("  %-"+fmt.Sprintf("%d", size+1)+"s %s\n", color.Notice.Sprint(key+")"), color.Secondary.Sprint(value))
}

// RenderChooseQuery is the method used by default input `Choose()` method to
// to render the query prompt choice (after the choices) into a string. Can be
// overwritten at users discretion.
var RenderChooseQuery = func() string {
	return "Choose: "
}

func (i *DefaultInput) Choose(question string, choices map[string]string, _default ...string) string {
	_default = append(_default, "")
	options := RenderChooseQuestion(question)
	keys := []string{}
	max := 0
	for k, _ := range choices {
		if l := len(k); l > max {
			max = l
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		options += RenderChooseOption(k, choices[k], max)
	}
	options += RenderChooseQuery()
	return i.Ask(options, func(in string) error {
		if _, ok := choices[in]; ok {
			return nil
		} else {
			return fmt.Errorf("Choose one of: %s", strings.Join(keys, ", "))
		}
	}, _default[0])
}

// ConfirmRejection is the message replied to the user if she does not answer
// with "yes", "y", "no" or "n" (case insensitive)
var ConfirmRejection = fmt.Sprintf("%s\n\n", color.Warn.Sprint(`Please respond with "yes" or "no"`))

// ConfirmYesRegex is the regular expression used to check if the user replied positive
var ConfirmYesRegex = regexp.MustCompile(`^(?i)y(es)?$`)

// ConfirmNoRegex is the regular expression used to check if the user replied negative
var ConfirmNoRegex = regexp.MustCompile(`^(?i)no?$`)

func (i *DefaultInput) Confirm(question string, _default ...bool) bool {
	_default = append(_default, false)
	cb := func(value string) error { return nil }
	for {
		res := i.Ask(question+color.Primary.Sprint(map[bool]string{true: " (Y|n) ", false: " (y|N) "}[_default[0]]), cb)
		if len(res) == 0 {
			return _default[0]
		}
		if ConfirmYesRegex.MatchString(res) {
			return true
		} else if ConfirmNoRegex.MatchString(res) {
			return false
		} else {
			i.out.Printf(ConfirmRejection)
		}
	}
}

func InputEmptyOk(s string) error {
	return nil
}

func InputAny(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("No input provided")
	} else {
		return nil
	}
}
