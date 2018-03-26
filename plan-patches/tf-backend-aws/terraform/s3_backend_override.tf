terraform {
  backend "s3" {
    bucket = "YOUR_BUCKET_NAME"
    key    = "YOUR_BBL_ENV_NAME/bbl-state/tf-state"
    region = "YOUR_BUCKET_REGION"
  }
}
