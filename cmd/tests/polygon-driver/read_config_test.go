package polygon_driver

import (
	"encoding/xml"
	"fmt"
	polygon_driver "github.com/Myriad-Dreamin/tarus/pkg/tarus-driver/polygon"
	"golang.org/x/net/html/charset"
	"strings"
	"testing"
)

func TestReadConfig(t *testing.T) {
	var cfg = `<?xml version="1.0" encoding="windows-1251" standalone="no"?>
<problem name="problem-name" revision="">
    <judging>
        <time-limit value="1000"/>
        <memory-limit value="268435456"/>
        <input-file value="stdin"/>
        <output-file value="stdout"/>
        <test-validator value="v.cpp"/>
        <checker auto-update="true" value="std::hcmp.cpp"/>
        <tests-well-formed value="true"/>
        <testlib auto-update="false"/>
    </judging>
    <packaging>
        <testset-pattern value="%s"/>
        <input-file-path-pattern value="%02d"/>
        <answer-file-path-pattern value="%02d.a"/>
        <statement-template-file value="problem.tex"/>
        <statement-path-pattern value="statements/%s/problem.tex"/>
        <checker-path value="check.exe"/>
        <render-formulas-using-mathjax value="true"/>
    </packaging>
</problem>`

	v := polygon_driver.ProblemConfig{}
	var dec = xml.NewDecoder(strings.NewReader(cfg))
	dec.CharsetReader = charset.NewReaderLabel
	err := dec.Decode(&v)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(v)
}
