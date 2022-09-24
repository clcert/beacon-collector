import hashlib
import sys
from to_txt import to_txt

filename = sys.argv[1]
with open(filename, 'r') as f:
  content = to_txt(f.read())
  hashObject = hashlib.sha3_512(content.encode('utf-8'))
  digest = hashObject.hexdigest()
print(digest)

