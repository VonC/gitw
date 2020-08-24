# gitw Git Wrapper

## Goal: force user to be identified in a Linux environment

When you are sharing the same service account in a Linux server, you want to make commit as "you", not "the service account".

This gitw.sh script will force the user to select their name/email at the first git command, then will reuse the same name/email

That means the following environment variable will always be set before any git command:

- `GIT_AUTHOR_NAME`
- `GIT_AUTHOR_EMAIL`
- `GIT_COMMITTER_NAME`
- `GIT_COMMITTER_EMAIL `

## Installation

### From Windows PC

- clone the repo, 
- launch `build.bat amd`
- copy the Linux executable `gitw` to the Linux server

### From the Linux server

- clone the repo
- put the gitw generate in the previous step 
- run once `gitw.sh gitwset`: that will add the alias for git in the `~<service account>/.bashrc` (or `.env` if you don't have the right to modify the `.bashrc`)
- source `.bashrc` (or `.env`)

### Configuration

copy `.gitusers.tpl` to `~/.gitusers`
Add a few names/emails in it (leavethe IP address to `0.0.0.0` at first)

## Usage

Try any git command, like `git version`

- the first command will trigger a popup menu asking to select a name
- once the name is selected, any subsequent command will be preceded by your name/email (just for information), and will use the right `user.name`/`user.email`.  
No need to re-select them as long as you are in the same shell session

Each new shell session would re-trigger the user name selection (once per session, at the first `git` command instance)