import json
import os
import pandas as pd

os.chdir('/Users/alexisdeviney/Documents/labs/lambdaCrackStation/')

with open('data/HashtoPlaintextDataMVP.json') as json_file:
    jsondata = json_file.read()
    lookupdict = json.loads(jsondata)


df = pd.DataFrame({"shaHash": key, "password": value} for (key, value) in lookupdict.items())
print(df.head())

df.to_csv("data/hashdata.csv.gz",index=False,compression='gzip')
