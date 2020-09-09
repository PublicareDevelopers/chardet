// Package chardet ports character set detection from ICU.
package chardet

import (
	"errors"
	"sort"
)

// Result contains all the information that charset detector gives.
type Result struct {
	// IANA name of the detected charset.
	Charset string
	// IANA name of the detected language. It may be empty for some charsets.
	Language string
	// Confidence of the Result. Scale from 1 to 100. The bigger, the more confident.
	Confidence int
}

// Detector implements charset detection.
type Detector struct {
	recognizers []recognizer
	stripTag    bool
}

// List of charset recognizers
var recognizers = []recognizer{
	newRecognizerUTF8(),
	newRecognizerUTF16be(),
	newRecognizerUTF16le(),
	newRecognizerUTF32be(),
	newRecognizerUTF32le(),
	newRecognizer8859m1en(),
	newRecognizer8859m1da(),
	newRecognizer8859m1de(),
	newRecognizer8859m1es(),
	newRecognizer8859m1fr(),
	newRecognizer8859m1it(),
	newRecognizer8859m1nl(),
	newRecognizer8859m1no(),
	newRecognizer8859m1pt(),
	newRecognizer8859m1sv(),
	newRecognizer8859m2cs(),
	newRecognizer8859m2hu(),
	newRecognizer8859m2pl(),
	newRecognizer8859m2ro(),
	newRecognizer8859m5ru(),
	newRecognizer8859m6ar(),
	newRecognizer8859m7el(),
	newRecognizer8859m8Ihe(),
	newRecognizer8859m8he(),
	newRecognizerWindows1251(),
	newRecognizerWindows1256(),
	newRecognizerKOI8R(),
	newRecognizer8859m9tr(),

	newRecognizersjis(),
	newRecognizerGB18030(),
	newRecognizereucJp(),
	newRecognizereucKr(),
	newRecognizerbig5(),

	newRecognizer2022JP(),
	newRecognizer2022KR(),
	newRecognizer2022CN(),

	newRecognizerIBM424HeRTL(),
	newRecognizerIBM424HeLTR(),
	newRecognizerIBM420ArRTL(),
	newRecognizerIBM420ArLTR(),
}

// NewTextDetector creates a Detector for plain text.
func NewTextDetector() *Detector {
	return &Detector{recognizers, false}
}

// NewHTMLDetector creates a Detector for HTML.
func NewHTMLDetector() *Detector {
	return &Detector{recognizers, true}
}

var (
	// ErrorNotDetected is error message if charset not detected
	ErrorNotDetected = errors.New("Charset not detected")
)

// DetectBest returns the Result with highest Confidence.
func (d *Detector) DetectBest(b []byte) (r *Result, err error) {
	var all []Result
	if all, err = d.DetectAll(b); err == nil {
		r = &all[0]
	}
	return
}

// DetectAll returns all Results which have non-zero Confidence. The Results are sorted by Confidence in descending order.
func (d *Detector) DetectAll(b []byte) ([]Result, error) {
	input := newRecognizerInput(b, d.stripTag)
	outputChan := make(chan recognizerOutput)
	for _, r := range d.recognizers {
		go matchHelper(r, input, outputChan)
	}
	outputs := make([]recognizerOutput, 0, len(d.recognizers))
	for i := 0; i < len(d.recognizers); i++ {
		o := <-outputChan
		if o.Confidence > 0 {
			outputs = append(outputs, o)
		}
	}
	if len(outputs) == 0 {
		return nil, ErrorNotDetected
	}

	sort.Sort(recognizerOutputs(outputs))
	dedupOutputs := make([]Result, 0, len(outputs))
	foundCharsets := make(map[string]struct{}, len(outputs))
	for _, o := range outputs {
		if _, found := foundCharsets[o.Charset]; !found {
			dedupOutputs = append(dedupOutputs, Result(o))
			foundCharsets[o.Charset] = struct{}{}
		}
	}
	if len(dedupOutputs) == 0 {
		return nil, ErrorNotDetected
	}
	return dedupOutputs, nil
}

func matchHelper(r recognizer, input *recognizerInput, outputChan chan<- recognizerOutput) {
	outputChan <- r.Match(input)
}

type recognizerOutputs []recognizerOutput

func (r recognizerOutputs) Len() int           { return len(r) }
func (r recognizerOutputs) Less(i, j int) bool { return r[i].Confidence > r[j].Confidence }
func (r recognizerOutputs) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
