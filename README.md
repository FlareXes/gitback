# GitBack

GitBack is a tool designed to backup GitHub *repositories*, *wikis*, and *gists* either with or without authentication. It provides the flexibility to backup public repositories without authentication or to backup both public and private repositories using a GitHub Personal Access Token (PAT).

## Features

- Backup repositories with or without authentication.
- Clone repositories and their wikis if available.
- Clone public and private gists (requires PAT).
- Supports concurrent downloads for faster backups.

## Dependencies

- [github.com/google/go-github/v59/github](https://pkg.go.dev/github.com/google/go-github/v59/github)

## Installation

1. Clone the GitBack repository:

```bash
git clone https://github.com/flarexes/gitback.git
```

2. Navigate to the cloned directory:

```bash
cd gitback
```

3. Build the project:

```bash
go build
```

4. (Optional) Set up a GitHub Personal Access Token (PAT) if you plan to backup private repositories. Instructions can be found [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).

5. Add your GitHub Personal Access Token as environment variable:

    ```
    GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
    ```

    - **Windows:**

        - Open Control Panel and navigate to System and Security > System > Advanced system settings > Environment Variables.
        - Under System Variables, click "New" and add a variable named GITHUB_PERSONAL_ACCESS_TOKEN with your PAT as the value.

    - **Linux (e.g., Ubuntu) & MacOS:**

        - Open your terminal and edit the .bashrc or .zshrc file using your preferred text editor (e.g., nano, vim, gedit):

        ```bash
        nano ~/.bashrc
        ```

        - Add the following line at the end of the file:

        ```bash
        export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
        ```

        - Save the file and exit. Then, reload the shell configuration:

        ```bash
        source ~/.bashrc
        ```

## Usage

```bash
./gitback [flags]
```

### Flags

| Flag        | Description                                                                                       | Required                  |
| ----------- | ------------------------------------------------------------------------------------------------- | ------------------------- |
| `-noauth`   | Disable GitHub authentication. Limits requests to 60 per hour and access to public data only.     | No                        |
| `-username` | Specify the GitHub username when using `-noauth`.                                                 | Yes (if `-noauth` is set) |
| `-thread`  | Set the maximum number of concurrent connections (default: 10).                                   | No                        |
| `-token`    | Provide the GitHub Personal Access Token directly as a flag (overrides the environment variable). | No                        |

### Folder Structure

When running the tool, the following folders will be created:

- **`gitback-backup_YYYY-MM-DD_HH-MM-SS/repos/`**: Contains cloned repositories and their wikis.
- **`gitback-backup_YYYY-MM-DD_HH-MM-SS/gists/`**: Contains cloned gists.

## Examples

### Backup Public Repositories (No Authentication)

```bash
./gitback -noauth -username flarexes
```

### Backup Private and Public Repositories (With Authentication)

```bash
./gitback
```

### Backup with Custom Thread Limit

```bash
./gitback -thread 20
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request, I want to take this project futher.

## Issues

If you encounter any issues or have suggestions for improvements, please open an issue on the [GitHub repository](https://github.com/flarexes/gitback/issues).


## License

This project is licensed under the BSD-3-Clause license. For more information, please see the [LICENSE](LICENSE) file.
