import hashlib
import sys
from parse_md import parse_md

filename = sys.argv[1]
with open(filename, 'r') as f:
  content = parse_md(f.read())
  hashObject = hashlib.sha3_512(content.encode('utf-8'))
  digest = hashObject.hexdigest()
print(digest)

