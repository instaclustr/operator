## Monitoring

Instaclustr allows you scrap clusters metrics via its Monitoring API.

If you use [Prometheus-Operator](https://github.com/prometheus-operator/prometheus-operator) as your monitoring system:

There are two ways of configuring Prometheus to scrap Instaclustr metrics:
1. By creating a configuration file and providing it to the Prometheus Operator by passing a secret with configuration of the job 

    A template of the job config:
    ```yaml
    - job_name: instaclustr_prometheus
      scheme: https
      http_sd_configs:
            - url: https://{ACCOUNT_ID}.prometheus.monitoring.dev.instaclustr.com/discovery/v1/
              basic_auth:
                #get from account settings page
                username: {USERNAME}
                password: {MONITORING_API_KEY}
      metrics_path: 'metrics/v2/query'
      basic_auth:
        #get from account settings page
        username: {USERNAME}
        password: {MONITORING_API_KEY}
    ```
    
    Creating a secret which contains the described config:
    ```shell
    kubectl create secret generic {your-secret-name} --from-file={your-config.yaml}
    ```
    
    By updating your prometheus with the following configuration will allow your prometheus to
    scrap metrics of your clusters on Instaclustr:
    ```yaml
    # Your prometheus manifest
      additionalScrapeConfigs:
        name: {your-secret-name}
        key: {your-config.yaml}
    ```

2. If your prometheus is deployed via Helm everything is getting easier.

    In your `values.yaml` file which stores the configuration of the whole kube-prometheus-stack just add the following
    settings to the configuration of prometheus:

    ```yaml
   additionalScrapeConfigs:
    - job_name: instaclustr_prometheus
      scheme: https
      http_sd_configs:
            - url: https://{ACCOUNT_ID}.prometheus.monitoring.dev.instaclustr.com/discovery/v1/
              basic_auth:
                #get from account settings page
                username: {USERNAME}
                password: {MONITORING_API_KEY}
      metrics_path: 'metrics/v2/query'
      basic_auth:
        #get from account settings page
        username: {USERNAME}
        password: {MONITORING_API_KEY}
      ```
   
    It will create a secret which stores the following jobs configuration and provide it to the prometheus.
    
If you want to scrap metrics from a specific cluster just provide its `clusterId` to the url:
```yaml
http_sd_configs:
  - url: https://{ACCOUNT_ID}.prometheus.monitoring.dev.instaclustr.com/discovery/v1/{CLUSTER_ID}
```

If you want to read more about Instaclustr Metrics and configuring of Prometheus please 
follow our [doc](https://www.instaclustr.com/support/api-integrations/integrations/instaclustr-monitoring-with-prometheus/). 