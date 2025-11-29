# Nest v1 Info

**Release Date**: TBD

## Subcommands

### Search

> *Searches the Arch repositories and AUR for a package*

```bash
# search all repos for `nest` package
nest search nest

# search the AUR and Extra repos for `nest` package
nest --repos=aur,extra search nest

# search all repos for `nest` package and return the top 5
nest --max-out=5 search nest
```

### Info

> *Prints the info for a package*

```bash
# print info for `nest` package
nest info nest

# print info for `nest` package only looking in the AUR and Extra repos
nest info --repos=aur,extra nest
```
