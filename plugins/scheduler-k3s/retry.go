package scheduler_k3s

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dokku/dokku/plugins/common"
	resty "github.com/go-resty/resty/v2"
)

// DefaultRetryAttempts is the number of times a network-dependent operation is
// attempted before giving up
const DefaultRetryAttempts = 5

// DefaultRetryInitialDelay is the delay before the first retry of a
// network-dependent operation. Each subsequent retry doubles the delay.
var DefaultRetryInitialDelay = 2 * time.Second

// RetryInput contains the input parameters for the retryOperation function
type RetryInput struct {
	// Description is a short description of the operation, used when logging failed attempts
	Description string

	// Attempts is the maximum number of attempts, defaulting to DefaultRetryAttempts
	Attempts int

	// InitialDelay is the delay before the first retry, defaulting to DefaultRetryInitialDelay
	InitialDelay time.Duration

	// Operation is the operation to attempt
	Operation func(ctx context.Context) error
}

// retryOperation runs an operation until it succeeds, retrying failed attempts
// with exponential backoff. The last error is returned once every attempt has
// failed, and the context error is returned if the context is cancelled while
// waiting between attempts.
func retryOperation(ctx context.Context, input RetryInput) error {
	attempts := input.Attempts
	if attempts <= 0 {
		attempts = DefaultRetryAttempts
	}

	delay := input.InitialDelay
	if delay <= 0 {
		delay = DefaultRetryInitialDelay
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = input.Operation(ctx)
		if lastErr == nil {
			return nil
		}

		if attempt == attempts {
			break
		}

		common.LogWarn(fmt.Sprintf("Attempt %d/%d to %s failed: %s, retrying in %s", attempt, attempts, input.Description, lastErr.Error(), delay))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}

	return fmt.Errorf("failed after %d attempts: %w", attempts, lastErr)
}

// retryExecCommand runs a local command, retrying it until it exits successfully
func retryExecCommand(ctx context.Context, description string, input common.ExecCommandInput) error {
	return retryOperation(ctx, RetryInput{
		Description: description,
		Operation: func(ctx context.Context) error {
			result, err := common.CallExecCommandWithContext(ctx, input)
			if err != nil {
				return err
			}
			if result.ExitCode != 0 {
				return fmt.Errorf("Invalid exit code: %d", result.ExitCode)
			}
			return nil
		},
	})
}

// retrySshCommand runs a command over ssh, retrying it until it exits successfully
func retrySshCommand(ctx context.Context, description string, input common.SshCommandInput) error {
	return retryOperation(ctx, RetryInput{
		Description: description,
		Operation: func(ctx context.Context) error {
			result, err := common.CallSshCommandWithContext(ctx, input)
			if err != nil {
				return err
			}
			if result.ExitCode != 0 {
				return fmt.Errorf("Invalid exit code: %d", result.ExitCode)
			}
			return nil
		},
	})
}

// DownloadToFileInput contains the input parameters for the downloadToFile function
type DownloadToFileInput struct {
	// Description is a short description of the download, used in log and error messages
	Description string

	// URL is the url to download
	URL string

	// Filename is the path the download is written to
	Filename string

	// Mode is the file mode of the written file
	Mode os.FileMode

	// Username is the owner of the written file
	Username string

	// GroupName is the group of the written file
	GroupName string
}

// downloadToFile downloads a url to a file, retrying the download until it
// returns a non-empty body with a 200 status code
func downloadToFile(ctx context.Context, input DownloadToFileInput) error {
	client := resty.New()
	var body string
	err := retryOperation(ctx, RetryInput{
		Description: fmt.Sprintf("download %s", input.Description),
		Operation: func(ctx context.Context) error {
			resp, err := client.R().
				SetContext(ctx).
				Get(input.URL)
			if err != nil {
				return err
			}
			if resp == nil {
				return fmt.Errorf("Missing response")
			}
			if resp.StatusCode() != 200 {
				return fmt.Errorf("Invalid status code: %d", resp.StatusCode())
			}
			if len(resp.Body()) == 0 {
				return fmt.Errorf("Empty response body")
			}

			body = resp.String()
			return nil
		},
	})
	if err != nil {
		return fmt.Errorf("Unable to download %s: %w", input.Description, err)
	}

	err = common.WriteStringToFile(common.WriteStringToFileInput{
		Content:   body,
		Filename:  input.Filename,
		GroupName: input.GroupName,
		Mode:      input.Mode,
		Username:  input.Username,
	})
	if err != nil {
		return fmt.Errorf("Unable to write %s to file: %w", input.Description, err)
	}

	fi, err := os.Stat(input.Filename)
	if err != nil {
		return fmt.Errorf("Unable to get %s file size: %w", input.Description, err)
	}

	if fi.Size() == 0 {
		return fmt.Errorf("Invalid %s filesize", input.Description)
	}

	return nil
}
