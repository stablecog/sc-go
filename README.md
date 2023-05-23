
# stablecog/sc-go

This is a multi-module workspace of various GO applications and packages used by [Stablecog](https://stablecog.com).

Reference the appendix section for an index of the various projects and repositories.
## Appendix

### Applications
These are the standalone applications in this repository, they have more details in their specific README.

- [Server](https://github.com/stablecog/sc-go/server) - The primary backend server and APIs.
- [Cron](https://github.com/stablecog/sc-go/cron) - Various cron jobs utilized by stablecog
- [Upload API](https://github.com/stablecog/sc-go/uploadapi) - API for user uploaded images

### Shared Modules
These are modules referenced by both applications.

- [Database](https://github.com/stablecog/sc-go/database) - Database interactions (SQL, Redis, etc.)
- [Shared](https://github.com/stablecog/sc-go/shared) - Shared models, constants, etc.
- [Utilities](https://github.com/stablecog/sc-go/utils) - General purpose utilities.

## Related

- [Stablecog App](https://github.com/stablecog/stablecog)
- [sc-worker](https://github.com/stablecog/sc-worker)

## Contributing

Contributions are always welcome!

To contribute, fork this repository, make your changes, and open a pull request.

This repository has a [VSCode devcontainer configuration](https://github.com/stablecog/sc-go/blob/master/.devcontainer/devcontainer.json), the easiest way to get started is to utilize this configuration with the [Dev Containers Extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers).

### Running

## Create mock data

In the dev container run the following to create some mock data:

```
cd server
go run . -load-mock-data
```

Due to the nature and complexity of the stablecog ecosystem, depending on custom cogs, third party APIs, S3 buckets, discord webhooks, and a multitude of other things - many dependent services need to be mocked in order to properly test changes locally.

Some things may not be possible to test or run locally easily, we are constantly making improvements and improving our test coverage + mocks to improve this sytem, so feel free to [create an issue](https://github.com/stablecog/sc-go) to help us track things we need to account for in local development.
