package prompt

import (
	"fmt"
	"io"
	"regexp"
)

func YesNo(msg string, r io.Reader, w io.Writer) bool {
	var re = regexp.MustCompile(`(?mi)^y(?:es)?$`)
	var resp string
	fmt.Fprintf(w, "%s [y/N]\n", msg)
	fmt.Fscanln(r, &resp)

	return re.MatchString(resp)
}
