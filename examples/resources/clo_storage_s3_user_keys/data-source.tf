resource "clo_storage_s3_user_keys" "s3_userkeys"{
  user_id = clo_storage_s3_user.s3_user.user_id
}
