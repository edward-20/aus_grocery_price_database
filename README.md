# Australian grocery price database (Travis)

https://auscost.com.au

This is an open database of grocery prices in Australia. Its goal is to track long-term price trends to help make good purchasing decisions and hold grocery stores to account for price increases.

The service reads grocery prices from Woolworths' and Coles' websites to an influxdb timeseries database.

In the future it could read from other Australian grocers, based on time, motivation, etc.

## Similar things (Travis)

The excellently-named [Heisse Preise](https://heisse-preise.io/) does what I'd like Auscost to do eventually. It's [open source](https://github.com/badlogic/heissepreise) and there's a good writeup on [Wired](https://www.wired.com/story/heisse-preise-food-prices/). I think there's much to be learned from its UI.

## Architecture

This service consists of three applications:

* aus_grocery_price_database
  * This application written in golang. Reads from grocery store web APIs and streams price data to the timeseries database.
* InfluxDB3 Cloud Instance
  * A timeseries database. Efficiently stores tagged numerical information, write-optimised and analytic optimised (ACID deprioritised).
* Custom Svelte Frontend (TBD)
  * a svelte frontend hitting the cloud influxdb3 instance

## Deployment Flow
1. With every non-trivial commit to github `main` (pull request, push) ensure the `VERSION` in `main.go` is updated. If you can tag it please do otherwise refer to step 2
2. Pull in the `main` branch and run `tag_and_deploy_release.sh`
3. Either manually or via github workflow (manually for now), create a docker image and push the image to your dockerhub repo.
4. Your deployment machine (vm, some server you own) will pull in the dockerhub image and redeploy the container. This can be manually or perhaps you could set up a cron job to examine if there's a change in the image.
5. Ensure that the deployment machine runs the docker image as `docker run -env--file .env image`, where the `.env` file contains all the environment variables required by `config` in `main.go`.

## Developer Workflow
TBD: previously just running it locally with `GO_ENV` and `.env.<GO_ENV>` and `.env` on WSL but perhaps should require a `.env` file to be loaded into the linux environment. The alternative is using docker for development and by virtue of that docker compose because docker alone can't watch for changes in dev files.

### Core Goals (Travis)

The primary goal of this project is to shift the balance of power in favour of consumers by presenting pricing information on groceries. Use cases include:

* Forecasting low prices based on periodicity
  * E.G. Navel oranges from Woolworths oscillate in price on a 2-week period
* Compare prices between stores
* Track long-term pricing trends
* Not being tricked by false "sales" where the sale price isn't any cheaper than the long-term trends

Another goal of this project is a low budget. This means minimal hosting and maintenance requirements. It needs to be simple and stable. Minimal external dependencies, both in terms of internal software stack and external services. Updating the current stack involves bumping two container version numbers and a few invocations of `fly deploy`.

### Further thoughts (Travis)
A significant frontend challenge is product differentiation. The data scraped from the grocer's storefronts varies from store to store. There is always a SKU ID, product name, department ID/name and price. The product name is quite dirty, repetitive and awkward. E.G. "40% Salt Reduced", or "Alva Baby Starry Sky Print Reusable Cloth Nappy", "Alva Baby Starry Night Print Reusable Cloth Nappy".

Woolworths provides a barcode number, coles does not. It would be great for a user to be able to scan a barcode on their phone and pull up the price history of that item.