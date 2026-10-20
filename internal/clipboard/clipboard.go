package clipboard

import (
	"errors"
	"os/exec"

	atotto "github.com/atotto/clipboard"
)

// Copy copies the given text to the system clipboard.
// It checks for standard clipboard utilities (wl-copy, xclip, xsel)
// or uses the atotto/clipboard package.
func Copy(text string) error {
	if text == "" {
		return errors.New("cannot copy empty text")
	}

	// Try atotto first
	err := atotto.WriteAll(text)
	if err == nil {
		return nil
	}

	// Fallback to explicit wl-copy if available
	if _, lookErr := exec.LookPath("wl-copy"); lookErr == nil {
		cmd := exec.Command("wl-copy")
		pipe, pipeErr := cmd.StdinPipe()
		if pipeErr == nil {
			if startErr := cmd.Start(); startErr == nil {
				_, _ = pipe.Write([]byte(text))
				_ = pipe.Close()
				if waitErr := cmd.Wait(); waitErr == nil {
					return nil
				}
			}
		}
	}

	// Fallback to xclip if available
	if _, lookErr := exec.LookPath("xclip"); lookErr == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		pipe, pipeErr := cmd.StdinPipe()
		if pipeErr == nil {
			if startErr := cmd.Start(); startErr == nil {
				_, _ = pipe.Write([]byte(text))
				_ = pipe.Close()
				if waitErr := cmd.Wait(); waitErr == nil {
					return nil
				}
			}
		}
	}

	return err
}
