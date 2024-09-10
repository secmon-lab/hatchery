# hatchery
A tool to gather data + logs from SaaS and save them to object storage

![overview](https://github.com/m-mizutani/hatchery/assets/605953/0d065e1e-1b40-493b-a9c5-8215f2e1691e)

## Motivation

Many SaaS services offer APIs for accessing data and logs, but managing them can be challenging due to various reasons:

- Audit logs are often set to expire after a few months.
- The built-in log search console provided by the service is not user-friendly and lacks centralized functionality for searching and analysis.

As a result, security administrators are required to gather logs from multiple services and store them in object storage for long-term retention and analysis. However, this process is complicated by the fact that each service has its own APIs and data formats, making it difficult to implement and maintain a tool to gather logs.

`hatchery` is a solution designed to address these challenges by collecting data and logs from SaaS services and storing them in object storage. This facilitates log retention and prepares the data for analysis by security administrators.

For those interested in importing logs from Cloud Storage to BigQuery, please refer to [swarm](https://github.com/m-mizutani/swarm).

## License

Apache License 2.0
