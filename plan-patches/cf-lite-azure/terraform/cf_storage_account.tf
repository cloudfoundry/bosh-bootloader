
output "blobstore_storage_account_name" {
  value="${azurerm_storage_account.bosh.name}"
}
output "blobstore_storage_access_key" {
  value="${azurerm_storage_account.bosh.primary_access_key}"
}