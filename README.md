Simple `gh` extension to `clone` and `cd` into a repository or 'repo create'.

# 🏃 Getting started
To start using it, just clone the repository, run
```bash
gh ext install dyne/gh-cd
```

Now, as soon as you run `gh cd [account] [repository]` the command will create a new repository at `github.com/account/repository` (if it doesn't exists), otherwise will clone it locally and `cd` into it.
If the repo is already cloned, it simply `cd` to the directory.

# 👷 Config
Configuratin happen through the file `~/.gitconfig`, in the section `gh-cd`. There are two possible keys:
- `basedir`: the absolute path in which we will clone the repository
- `protocol`: which can be `ssh` or `https`
- `create-repo`: create repo if it doesn't exist
- `shell-cmd`: shell to be run in the directory repo
