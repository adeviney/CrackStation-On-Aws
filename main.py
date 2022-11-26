import boto3
import hashlib

s3 = boto3.client('s3')

BUCKET_NAME = 'crackstation-data'
FILE_NAME = 'hashdata.csv.gz'

def getPasswordByQueryS3(shaHash):
    resp = s3.select_object_content(
        Bucket=BUCKET_NAME,
        Key=FILE_NAME,
        ExpressionType='SQL',
        Expression=f"SELECT password FROM S3object s where s.shaHash = '{shaHash}'",
        InputSerialization = {'CSV': {"FileHeaderInfo": "Use"}, 'CompressionType': 'GZIP'},
        OutputSerialization = {'CSV': {}},
    )

    for event in resp['Payload']:
        if 'Records' in event:
            records = event['Records']['Payload'].decode('utf-8')
            return records
        elif 'Stats' in event:
            statsDetails = event['Stats']['Details']
            print("Stats details bytesScanned: ")
            print(statsDetails['BytesScanned'])
            print("Stats details bytesProcessed: ")
            print(statsDetails['BytesProcessed'])
            print("Stats details bytesReturned: ")
            print(statsDetails['BytesReturned'])

hashedPassword = hashlib.sha1(b"a").hexdigest()
crackedPassword = getPasswordByQueryS3(hashedPassword)
print(crackedPassword)


def lambda_handler(event, context):
    # TODO implement
    return {
        'statusCode': 200,
        'body': json.dumps('Hello from Lambda!')
    }
