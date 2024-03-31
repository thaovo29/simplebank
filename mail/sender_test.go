package mail

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thaovo29/simplebank/util"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	config, err := util.LoadConfig("..")
	require.NoError(t, err)
	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test gmail"
	content := `
	<h1> Hello world! </h1>
	<p>This is a test email</p>
	`

	to := []string {
		"thaovo2962002@gmail.com",
	}

	attachedFiles := []string {"../README.md"}
	err = sender.SendEmail(subject, content, to, nil, nil, attachedFiles)
	require.NoError(t, err)
}
