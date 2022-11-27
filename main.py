from lambda_function_s3 import *
import hashlib

shaHash = hashlib.sha1(b'a').hexdigest()
endpoint_request_body = {
    "shaHash": f"{shaHash}",
    "context": {}
}


response_from_lambda = lambda_handler(event = endpoint_request_body, context = endpoint_request_body['context'])
print(response_from_lambda)