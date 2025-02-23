package redirection

import (
    "testing"
)

func TestParseRedirection(t *testing.T) {
    tests := []struct {
        name      string
        args      []string
        wantArgs  []string
        wantRedir *Redirection
        wantErr   bool
    }{
        {
            name:      "no redirection",
            args:      []string{"arg1", "arg2"},
            wantArgs:  []string{"arg1", "arg2"},
            wantRedir: nil,
            wantErr:   false,
        },
        {
            name:     "output redirection",
            args:     []string{"arg1", ">", "output.txt"},
            wantArgs: []string{"arg1"},
            wantRedir: &Redirection{
                Type: OutputRedirect,
                File: "output.txt",
            },
            wantErr: false,
        },
        {
            name:     "output append",
            args:     []string{"arg1", ">>", "output.txt"},
            wantArgs: []string{"arg1"},
            wantRedir: &Redirection{
                Type: OutputAppend,
                File: "output.txt",
            },
            wantErr: false,
        },
        {
            name:     "error redirection",
            args:     []string{"arg1", "2>", "error.txt"},
            wantArgs: []string{"arg1"},
            wantRedir: &Redirection{
                Type: ErrorRedirect,
                File: "error.txt",
            },
            wantErr: false,
        },
        {
            name:    "missing file",
            args:    []string{"arg1", ">"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            gotArgs, gotRedir, err := ParseRedirection(tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseRedirection() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr {
                if !sliceEqual(gotArgs, tt.wantArgs) {
                    t.Errorf("ParseRedirection() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
                }
                if !redirectionEqual(gotRedir, tt.wantRedir) {
                    t.Errorf("ParseRedirection() gotRedir = %v, want %v", gotRedir, tt.wantRedir)
                }
            }
        })
    }
}

func sliceEqual(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func redirectionEqual(a, b *Redirection) bool {
    if a == nil && b == nil {
        return true
    }
    if a == nil || b == nil {
        return false
    }
    return a.Type == b.Type && a.File == b.File
}