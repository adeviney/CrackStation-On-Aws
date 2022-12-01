import boto3
import json

def lambda_handler(event, context):
    shaHash = event['shaHash']
    s3 = boto3.client('s3')

    BUCKET_NAME = 'crackstation-data'
    FILE_NAME = 'hashdata.csv.gz'

    resp = s3.select_object_content(
        Bucket=BUCKET_NAME,
        Key=FILE_NAME,
        ExpressionType='SQL',
        Expression=f"SELECT s.password FROM S3object s where s.shaHash = '{shaHash}'",
        InputSerialization = {'CSV': {"FileHeaderInfo": "Use"}, 'CompressionType': 'GZIP'},
        OutputSerialization = {'CSV': {}}
    )

    # This is the event stream in the response
    event_stream = resp['Payload']
    end_event_received = False
    passwords = None

    # Iterate over events in the event stream as they come
    for e in event_stream:
        # If we received a records event, write the data to a file
        if 'Records' in e:
            passwords = (e['Records']['Payload'].decode('utf-8'))
            password = passwords.strip()
            responseString = f'{{ "{shaHash}": "{password}" }}'
            print(responseString)
            responseString = json.loads(responseString)
        # If we received a progress event, print the details
        elif 'Progress' in e:
            print(e['Progress']['Details'])
        # End event indicates that the request finished successfully
        elif 'End' in e:
            print('Result is complete')
            end_event_received = True

    if not end_event_received:
        raise Exception("Incomplete: End event not received, request incomplete.")
    elif not passwords:
        raise Exception(f"Uncrackable: CrackStation could not crack this hash, {shaHash}. Either it is not known to CrackStation or it is salted.")
    
    return responseString
