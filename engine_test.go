package arry

import (
    "bytes"
    "testing"
    "io/ioutil"
)


func TestEngine(t *testing.T) {
    file, _ := ioutil.ReadFile("_example/assets/static.html")

    engine := NewEngine("_example/assets/", "html")

    buf := new(bytes.Buffer)
    engine.Render(buf, "static.html", nil, nil)

    if buf.String() != string(file) {
        t.Errorf("engine rendering is not correct, %s", buf.String())
    }
}
