import sys

def to_txt(content):
  return content.replace('**', '').replace('#', '').replace('[','<').replace(']','>')

if __name__ == '__main__':
  filename = sys.argv[1]
  with open(filename, 'r') as f:
    content = f.read()
    print(to_txt(content))