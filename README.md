Simple `gh` extension to `clone` and `cd` into a repository.

# 🏃 Getting started
To start using it, just clone the repository, run `go build` and then `gh extension install .`

Now, as soon as you run `gh cd [account] [repository]` the command will create a new repository at `github.com/account/repository` (if it doesn't exists), clone it locally and `cd` into it.

# 👷 Config
Configuratin happen through the file `~/.gitconfig`, in the section `gh-cd`. There are two possible keys:
- `basedir`: the path in which we will clone the repository
- `protocol`: which can be `ssh` or `https`
