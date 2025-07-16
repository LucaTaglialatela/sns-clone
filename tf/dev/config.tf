terraform {
  backend "s3" {
    bucket  = "intern-luca"
    key     = "terraform.tfstate"
    region  = "ap-northeast-1"
    encrypt = "true"
  }
}

provider "aws" {
  region = "ap-northeast-1"
  default_tags {
    tags = {
      Project = "intern-luca"
    }
  }
}
