import boto3
import os

def lambda_handler(event, context):
    s3 = boto3.client('s3')
    bucket = os.environ.get('BUCKET_NAME')
    if bucket:
        for obj in s3.list_objects_v2(Bucket=bucket).get('Contents', []):
            s3.delete_object(Bucket=bucket, Key=obj['Key'])
    return {'status': 'ok'}
