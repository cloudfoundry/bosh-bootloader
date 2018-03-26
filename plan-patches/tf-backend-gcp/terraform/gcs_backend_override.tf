terraform {
  backend "gcs" {
    bucket      = "YOUR_BUCKET_NAME"
    prefix      = "YOUR_BUCKET_PREFIX"
    credentials = "YOUR_GCP_SERVICE_ACCOUNT_KEY_PATH"
  }
}
