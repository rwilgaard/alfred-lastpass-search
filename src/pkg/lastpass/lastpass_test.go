package lastpass

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// TestHelperProcess isn't a real test. It's used as a helper process
// to simulate the `lpass` binary.
// It's invoked by tests that replace `ExecCommand`.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	mockExitCode := os.Getenv("MOCK_EXIT_CODE")
	mockStdout := os.Getenv("MOCK_STDOUT")
	mockStderr := os.Getenv("MOCK_STDERR")

	if mockStdout != "" {
		fmt.Fprint(os.Stdout, mockStdout)
	}
	if mockStderr != "" {
		fmt.Fprint(os.Stderr, mockStderr)
	}

	if mockExitCode == "0" {
		os.Exit(0)
	} else {
		exitCode := 1
		if parsedCode, err := fmt.Sscan(mockExitCode, &exitCode); err != nil || parsedCode != 1 {
		}
		os.Exit(exitCode)
	}
}

// mockExecCommand returns a function that, when called, returns an *exec.Cmd
// configured to run TestHelperProcess. It sets environment variables to control
// TestHelperProcess's behavior.
func mockExecCommand(t *testing.T, stdoutData string, stderrData string, exitCode int) func(string, ...string) *exec.Cmd {
	t.Helper()
	return func(cmdPath string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", cmdPath}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			fmt.Sprintf("MOCK_STDOUT=%s", stdoutData),
			fmt.Sprintf("MOCK_STDERR=%s", stderrData),
			fmt.Sprintf("MOCK_EXIT_CODE=%d", exitCode),
		}
		return cmd
	}
}

func TestLastpassServiceIsLoggedIn(t *testing.T) {
	testCases := []struct {
		name         string
		mockCmdName  string
		mockCmdArgs  string
		mockStdout   string
		mockStderr   string
		mockExitCode int
		binPath      string
		wantLoggedIn bool
	}{
		{
			name:         "Logged In",
			mockCmdName:  "lpass",
			mockCmdArgs:  "status --quiet",
			mockStdout:   "",
			mockStderr:   "",
			mockExitCode: 0,
			binPath:      "lpass",
			wantLoggedIn: true,
		},
		{
			name:         "Not Logged In",
			mockCmdName:  "lpass",
			mockCmdArgs:  "status --quiet",
			mockStdout:   "",
			mockStderr:   "Not logged in",
			mockExitCode: 1,
			binPath:      "lpass",
			wantLoggedIn: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ls := &LastpassService{
				BinPath:     tt.binPath,
				ExecCommand: mockExecCommand(t, tt.mockStdout, tt.mockStderr, tt.mockExitCode),
			}
			if got := ls.IsLoggedIn(); got != tt.wantLoggedIn {
				t.Errorf("IsLoggedIn() = %v, want %v", got, tt.wantLoggedIn)
			}
		})
	}
}

func TestLastpassServiceGetFolders(t *testing.T) {
	testCases := []struct {
		name            string
		mockStdout      string
		mockStderr      string
		mockExitCode    int
		wantFolders     []LastpassFolder
		wantErr         bool
		expectedErrText string
	}{
		{
			name:         "No folders",
			mockStdout:   "",
			mockStderr:   "",
			mockExitCode: 0,
			wantFolders:  []LastpassFolder{},
			wantErr:      false,
		},
		{
			name:         "Single folder",
			mockStdout:   "Work/\n",
			mockStderr:   "",
			mockExitCode: 0,
			wantFolders: []LastpassFolder{
				{Name: "Work/"},
			},
			wantErr: false,
		},
		{
			name:         "Multiple folders",
			mockStdout:   "Personal/\nShared-Family/\nWork/\n",
			mockStderr:   "",
			mockExitCode: 0,
			wantFolders: []LastpassFolder{
				{Name: "Personal/"},
				{Name: "Shared-Family/"},
				{Name: "Work/"},
			},
			wantErr: false,
		},
		{
			name:         "Folders with spaces and special characters",
			mockStdout:   "Finance/Banking/\nOld Projects/\nWork Items (New)/\n",
			mockStderr:   "",
			mockExitCode: 0,
			wantFolders: []LastpassFolder{
				{Name: "Finance/Banking/"},
				{Name: "Old Projects/"},
				{Name: "Work Items (New)/"},
			},
			wantErr: false,
		},
		{
			name:         "Output with trailing newline already handled by TrimSuffix",
			mockStdout:   "Folder1/\nFolder2/\n",
			mockStderr:   "",
			mockExitCode: 0,
			wantFolders: []LastpassFolder{
				{Name: "Folder1/"},
				{Name: "Folder2/"},
			},
			wantErr: false,
		},
		{
			name:            "lpass command fails",
			mockStdout:      "",
			mockStderr:      "lpass command not found",
			mockExitCode:    127,
			wantFolders:     nil,
			wantErr:         true,
			expectedErrText: "error running command to get folders",
		},
		{
			name:            "lpass ls returns error output",
			mockStdout:      "",
			mockStderr:      "Error: Not logged in to LastPass.",
			mockExitCode:    1,
			wantFolders:     nil,
			wantErr:         true,
			expectedErrText: "error running command to get folders",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ls := &LastpassService{
				BinPath:     "lpass",
				ExecCommand: mockExecCommand(t, tt.mockStdout, tt.mockStderr, tt.mockExitCode),
			}

			gotFolders, err := ls.GetFolders()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetFolders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.expectedErrText != "" && (err == nil || !strings.Contains(err.Error(), tt.expectedErrText)) {
					t.Errorf("GetFolders() error = %v, want error containing %q", err, tt.expectedErrText)
				}
				if !reflect.DeepEqual(gotFolders, tt.wantFolders) {
					t.Errorf("GetFolders() gotFolders = %v, want %v (on error path)", gotFolders, tt.wantFolders)
				}
				return
			}

			if !reflect.DeepEqual(gotFolders, tt.wantFolders) {
				t.Errorf("GetFolders() gotFolders = %v, want %v", gotFolders, tt.wantFolders)
			}
		})
	}
}

func TestLastpassServiceGetEntries(t *testing.T) {
	type fields struct {
		BinPath string
	}
	type args struct {
		query   string
		folders []string
		fuzzy   bool
	}
	testCases := []struct {
		name            string
		fields          fields
		args            args
		mockStdout      string
		mockStderr      string
		mockExitCode    int
		want            []LastpassEntry
		wantErr         bool
		expectedErrText string
	}{
		{
			name: "Test with empty query and folders - no entries",
			fields: fields{
				BinPath: "lpass",
			},
			args: args{
				query:   "",
				folders: []string{},
				fuzzy:   false,
			},
			mockStdout:   "",
			mockStderr:   "",
			mockExitCode: 0,
			want:         []LastpassEntry{},
			wantErr:      false,
		},
		{
			name: "One entry, no query, default folder",
			fields: fields{
				BinPath: "lpass",
			},
			args: args{
				query:   "",
				folders: []string{},
				fuzzy:   false,
			},
			mockStdout:   "Work/My Entry [id: 123] [url: http://example.com] [username: user1] pass1\n",
			mockStderr:   "",
			mockExitCode: 0,
			want: []LastpassEntry{
				{ID: "123", Name: "My Entry", Folder: "Work", URL: "http://example.com", Username: "user1", Password: "pass1"},
			},
			wantErr: false,
		},
		{
			name: "Two entries, query matches one",
			fields: fields{
				BinPath: "lpass",
			},
			args: args{
				query:   "ServiceA",
				folders: []string{},
				fuzzy:   false,
			},
			mockStdout: "Dev/ServiceA [id: 100] [url: http://service-a.com] [username: dev_a] pass_a\n" +
				"Prod/ServiceB [id: 101] [url: http://service-b.com] [username: prod_b] pass_b\n",
			mockStderr:   "",
			mockExitCode: 0,
			want: []LastpassEntry{
				{ID: "100", Name: "ServiceA", Folder: "Dev", URL: "http://service-a.com", Username: "dev_a", Password: "pass_a"},
			},
			wantErr: false,
		},
		{
			name: "Specific folder, one entry",
			fields: fields{
				BinPath: "lpass",
			},
			args: args{
				query:   "",
				folders: []string{"Social"},
				fuzzy:   false,
			},
			mockStdout:   "Social/Twitter [id: 789] [url: http://twitter.com] [username: mytwitter] twpass\n",
			mockStderr:   "",
			mockExitCode: 0,
			want: []LastpassEntry{
				{ID: "789", Name: "Twitter", Folder: "Social", URL: "http://twitter.com", Username: "mytwitter", Password: "twpass"},
			},
			wantErr: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ls := &LastpassService{
				BinPath: tt.fields.BinPath,
				// The mockExecCommand needs to be smart enough to handle the folder argument
				// that GetEntries appends to the command.
				// The `expectedCmdArgs` in mockExecCommand should be the *base* command.
				// TestHelperProcess will receive the full command including the folder.
				ExecCommand: mockExecCommand(t, tt.mockStdout, tt.mockStderr, tt.mockExitCode),
			}
			got, err := ls.GetEntries(tt.args.query, tt.args.folders, tt.args.fuzzy)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.expectedErrText != "" && (err == nil || !strings.Contains(err.Error(), tt.expectedErrText)) {
					t.Errorf("GetEntries() error = %v, want error containing %v", err, tt.expectedErrText)
				}
				// If error is expected, we might not need to check `got` further, or check if it's nil/empty
				if !reflect.DeepEqual(got, tt.want) { // tt.want should be nil or empty for error cases
					t.Errorf("GetEntries() got = %v, want %v (on error path)", got, tt.want)
				}
				return // Important to return after handling error case
			}

			// If no error was wanted, proceed with deep equal check
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEntries() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastpassServiceGetDetails(t *testing.T) {
	testCases := []struct {
		name            string
		itemID          string
		mockStdout      string
		mockStderr      string
		mockExitCode    int
		wantKeys        []string
		wantDetails     map[string]string
		wantErr         bool
		expectedErrText string
	}{
		{
			name:   "Successful retrieval",
			itemID: "12345",
			mockStdout: `SiteNameOrPath [id: 12345]
Username: user@example.com
Password: securepassword
URL: https://example.com
Notes: This is a note.
More notes on next line, but will be missed by current parser.
Last Modified: Thu Jan 01 00:00:00 UTC 1970
`,
			mockStderr:   "",
			mockExitCode: 0,
			wantKeys: []string{
				"Username",
				"Password",
				"URL",
				"Notes",
			},
			wantDetails: map[string]string{
				"Username": "user@example.com",
				"Password": "securepassword",
				"URL":      "https://example.com",
				"Notes":    "This is a note.",
			},
			wantErr: false,
		},
		{
			name:            "Item not found - command error",
			itemID:          "nonexistent",
			mockStdout:      "",
			mockStderr:      "Error: Could not find item 'nonexistent'",
			mockExitCode:    1,
			wantKeys:        nil,
			wantDetails:     nil,
			wantErr:         true,
			expectedErrText: "error running lpass show for itemID 'nonexistent'",
		},
		{
			name:            "Empty itemID",
			itemID:          "",
			mockStdout:      "", // Command won't be called
			mockStderr:      "",
			mockExitCode:    0,
			wantKeys:        nil,
			wantDetails:     nil,
			wantErr:         true,
			expectedErrText: "itemID is empty",
		},
		{
			name:         "Only header line from lpass",
			itemID:       "67890",
			mockStdout:   "OnlyHeader [id: 67890]\n",
			mockStderr:   "",
			mockExitCode: 0,
			wantKeys:     []string{},
			wantDetails:  map[string]string{},
			wantErr:      false,
		},
		{
			name:   "No notes field, parsing completes",
			itemID: "11223",
			mockStdout: `AnotherSite [id: 11223]
URL: http://anothersite.com
Username: anotheruser
`, // No "Notes:" field
			mockStderr:   "",
			mockExitCode: 0,
			wantKeys: []string{
				"URL",
				"Username",
			},
			wantDetails: map[string]string{
				"URL":      "http://anothersite.com",
				"Username": "anotheruser",
			},
			wantErr: false,
		},
		{
			name:   "Field with empty value",
			itemID: "33445",
			mockStdout: `ItemWithEmptyField [id: 33445]
URL: http://site.com
CustomField: 
Notes: Some notes.
`,
			mockStderr:   "",
			mockExitCode: 0,
			wantKeys: []string{
				"URL",
				"CustomField",
				"Notes",
			},
			wantDetails: map[string]string{
				"URL":         "http://site.com",
				"CustomField": "",
				"Notes":       "Some notes.",
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.itemID == "" && tt.wantErr && tt.expectedErrText == "itemID is empty" {
				ls := &LastpassService{BinPath: "lpass"}
				_, _, err := ls.GetDetails(tt.itemID)
				if err == nil {
					t.Fatalf("GetDetails() with empty itemID expected an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedErrText) {
					t.Errorf("GetDetails() with empty itemID error = %v, want error containing %q", err, tt.expectedErrText)
				}
				return
			}

			ls := &LastpassService{
				BinPath:     "lpass",
				ExecCommand: mockExecCommand(t, tt.mockStdout, tt.mockStderr, tt.mockExitCode),
			}

			gotKeys, gotDetails, err := ls.GetDetails(tt.itemID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.expectedErrText != "" && (err == nil || !strings.Contains(err.Error(), tt.expectedErrText)) {
					t.Errorf("GetDetails() error = %v, want error containing %q", err, tt.expectedErrText)
				}
				if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
					t.Errorf("GetDetails() gotKeys = %v, want %v (on error path)", gotKeys, tt.wantKeys)
				}
				if !reflect.DeepEqual(gotDetails, tt.wantDetails) {
					t.Errorf("GetDetails() gotDetails = %v, want %v (on error path)", gotDetails, tt.wantDetails)
				}
				return
			}

			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("GetDetails() gotKeys = %v, want %v", gotKeys, tt.wantKeys)
			}
			if !reflect.DeepEqual(gotDetails, tt.wantDetails) {
				t.Errorf("GetDetails() gotDetails = %v, want %v", gotDetails, tt.wantDetails)
			}
		})
	}
}
