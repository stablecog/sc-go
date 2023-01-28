
# stablecog/go-apps

This is a multi-module workspace of various GO applications and packages used by [Stablecog](https://stablecog.com).

Reference the appendix section for an index of the various projects and repositories.
## Appendix

### Applications
These are the standalone applications in this repository, they have more details in their specific README.

- [Server](https://github.com/stablecog/go-apps/server) - The primary backend server and APIs.
- [Cron](https://github.com/stablecog/go-apps/cron) - Various cron jobs utilized by stablecog

### Shared Modules
These are modules referenced by both applications.

- [Database](https://github.com/stablecog/go-apps/database) - Database interactions (SQL, Redis, etc.)
- [Shared](https://github.com/stablecog/go-apps/shared) - Shared models, constants, etc.
- [Utilities](https://github.com/stablecog/go-apps/utils) - General purpose utilities.

## Related

- [Stablecog App](https://github.com/stablecog/stablecog)

## Contributing

Contributions are always welcome!

To contribute, fork this repository, make your changes, and open a pull request.

This repository has a [VSCode devcontainer configuration](https://github.com/stablecog/go-apps/blob/master/.devcontainer/devcontainer.json), the easiest way to get started is to utilize this configuration with the [Dev Containers Extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers).

### Running
Due to the nature and complexity of the stablecog ecosystem, depending on custom cogs, third party APIs, S3 buckets, discord webhooks, and a multitude of other things - many dependent services need to be mocked in order to properly test changes locally.

Some things may not be possible to test or run locally easily, we are constantly making improvements and improving our test coverage + mocks to improve this sytem, so feel free to [create an issue](https://github.com/stablecog/go-apps) to help us track things we need to account for in local development.
