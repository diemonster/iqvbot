variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table to create"
  default     = "iqvbot"
}

variable "dynamodb_read_capacity" {
  description = "Read capacity for the DynamoDB table"
  default     = 10
}

variable "dynamodb_write_capacity" {
  description = "Write capacity for the DynamoDB table"
  default     = 10
}

variable "iam_user_name" {
  description = "Name of the IAM user to create"
  default     = "iqvbot"
}

variable "docker_image" {
  description = "Docker image to run"
  default     = "quintilesims/iqvbot:latest"
}

variable "slack_bot_token" {
  description = "Authentication token for the Slack bot"
}

variable "slack_app_token" {
  description = "Authentication token for the Slack app"
}

variable "tenor_key" {
  description = "Authentication token for Tenor"
}

variable "deploy_name" {
  description = "Name of the Layer0 deploy to create"
  default     = "iqvbot"
}

variable "service_name" {
  description = "Name of the Layer0 service to create"
  default     = "iqvbot"
}

variable "environment_id" {
  description = "ID of the Layer0 environment to build the service"
}
