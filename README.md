# Recipes

[![codecov](https://codecov.io/gh/cdriehuys/recipes/graph/badge.svg?token=5pKiYfFX59)](https://codecov.io/gh/cdriehuys/recipes)

A web app written in Go to organize recipes.

## Development

Development is done through a dev container running in VS Code. This creates an
isolated container-based environment with all the tools necessary for
development.

### Dev Server

The server is run through the provided VS Code launch configuration which allows
for debugging through the IDE.

### Other Tools

[Just][just] is used to run almost all tasks related to development. For a full
list of available commands, run:
```shell
just --list
```

#### Live CSS Rebuilding

Styling uses [Tailwind CSS][tailwind-css] which depends on a build process. To
automatically rebuild the CSS file on any changes, use:
```shell
just dev
```

#### Database Migrations

Database migrations are executed through [`tern`][tern]. There are a few tasks
related to migrations:
```shell
# Migrate the database to the latest version:
just migrate

# Migrate to a specific version, eg undo the last applied migration:
just migrate-to -1

# Create a new migration with the correct numbering scheme:
just new-migration create_my_table
```

[just]: https://github.com/casey/just
[tailwind-css]: https://tailwindcss.com/
[tern]: https://github.com/jackc/tern/
