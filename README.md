## Flynn Cluster Backup

This service performs a backup of the flynn cluster on the hour.

The backups are stored on s3 and cleaned by hourly and daily counts.

### Config
|Env Vars|
|---|
|FLYNN_CLUSTER_PIN|
|FLYNN_URL|
|FLYNN_TOKEN|
|AWS_ACCESS_KEY_ID|
|AWS_SECRET_ACCESS_KEY|
|BACKUP_BUCKET|
|MAX_HOURLY|
|MAX_DAILY|


### Dev

#### Build

`make docker`

#### Run

```
docker run --rm --env-file ./env_file flynn-backup
```
