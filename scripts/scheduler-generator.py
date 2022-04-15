import argparse

prefix = "Sched"


def specToStr(i, X) -> str:
  specL = []
  specL.extend(["Sched", str(i), "x"])
  specL.extend(["{:03d}".format(k) for k in X])
  return ''.join(specL)

class Sched:

  def __init__(self, N:int):
    self.N = N
    self.visited = set()

  def generate(self, i, X):
    key = "{}-{}".format(i, X)
    if key in self.visited:
      return

    self.visited.add(key)
    #print(key)
    # print out the line
    specL = []
    specL.append(specToStr(i,X))
    specL.append(" = ")

    if i in X:
      for j in X:
        XX = X.copy()
        XX.remove(j)
        specL.append("b{:03d}(x).".format(j))
        specL.append(specToStr(i,XX))
        specL.append(" + ")
        self.generate(i, XX)
      specL = specL[:-1]
    else:
      for j in X:
        XX = X.copy()
        XX.remove(j)
        specL.append("b{:03d}(x).".format(j))
        specL.append(specToStr(i,XX))
        specL.append(" + ")
        self.generate(i, XX)
      specL.append("a{:03d}(x).".format(i))
      XX = X.copy()
      XX.add(i)
      ii = (i+1)%self.N
      specL.append(specToStr(ii,XX))
      self.generate(ii, XX)
      #specL = specL[:-1]

    spec = ''.join(specL)
    print(spec)


def main():
  parser = argparse.ArgumentParser()
  parser.add_argument("--N", type=int)
  args = parser.parse_args()
  N = args.N
  S = Sched(N)
  print("Scheduler = Sched0x")
  S.generate(0, set())
  print("Scheduler")

if __name__ == "__main__":
  main()