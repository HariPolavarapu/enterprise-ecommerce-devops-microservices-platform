resource "aws_s3_bucket" "logs" {
  bucket = "${var.env}-platform-logs"
}

resource "aws_cloudwatch_log_group" "app" {
  name = "/aws/eks/${var.env}/app"
}
