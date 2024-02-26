# GitBack

GitBack is a tool designed to backup GitHub repositories either with or without authentication. It provides the flexibility to backup public repositories without authentication or to backup both public and private repositories using a GitHub Personal Access Token (PAT).

## Dependencies

- [github.com/google/go-github/v59/github](https://pkg.go.dev/github.com/google/go-github/v59/github)
- [github.com/joho/godotenv](https://pkg.go.dev/github.com/joho/godotenv)

## Installation

1. Clone the GitBack repository:

```bash
git clone https://github.com/flarexes/gitback.git
```

2. Navigate to the cloned directory:

```bash
cd gitback/gitback
```

3. Build the project:

```bash
go build
```

4. Set up a GitHub Personal Access Token (PAT) if you plan to backup private repositories. Instructions can be found [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).

5. Create a `.env` file in the root directory of the project and add your GitHub Personal Access Token:

    ```
    GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
    ```

    Alternatively, you can set the environment variable directly on your system:

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

- `-noauth`: Disable GitHub authentication. Limited to 60 requests per hour and only public data can be accessed.
- `-username`: Required when `--noauth` flag is set. Specify the GitHub username to backup public repositories.

## Examples

### Backup Public Repositories (No Authentication)

```bash
./gitback -noauth -username flarexes
```

### Backup Private and Public Repositories (With Authentication)

```bash
./gitback
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request, I want to take this project futher.

## Issues

If you encounter any issues or have suggestions for improvements, please open an issue on the [GitHub repository](https://github.com/flarexes/gitback/issues).


## License

This project is licensed under the BSD-3-Clause license. For more information, please see the [LICENSE](LICENSE) file.
