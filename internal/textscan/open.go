package textscan

import (
	"os"

	"wc-stats-mcp/internal/validate"
)

// OpenFile opens a regular file after ResolveUnderRoot. The caller must Close the file.
func OpenFile(root, userPath string, lim validate.Limits) (f *os.File, resolved string, err error) {
	resolved, _, err = validate.ResolveUnderRoot(root, userPath, lim)
	if err != nil {
		return nil, "", err
	}
	f, err = os.Open(resolved)
	if err != nil {
		return nil, "", err
	}
	return f, resolved, nil
}
