import argparse


class Cycler:
  def __init__(self, N : int):
    self.N = N

  def genArgs(self, i) -> str:
    return "a{:03d},b{:03d},c{:03d},c{:03d}".format(i,i,i, (i-1)%self.N)

  def generate(self):
    specL = ["S = "]
    for i in range(self.N):
      specL.append("$c{:03d}.".format(i))
    specL.append("(")
    specL.extend(["A(", self.genArgs(0), ")"])

    for i in range(1, self.N):
      specL.extend(["| D(", self.genArgs(i), ")"])
    specL.append(")")
    spec = ''.join(specL)
    print(spec)

def main():
  parser = argparse.ArgumentParser()
  parser.add_argument("--N", type=int)
  args = parser.parse_args()
  N = args.N
  C = Cycler(N)
  print("A(a,b,c,d) = a(x).C(a,b,c,d)")
  print("B(a,b,c,d) = b(x).A(a,b,c,d)")
  print("C(a,b,c,d) = c(x).E(a,b,c,d)")
  print("D(a,b,c,d) = d'<d>.A(a,b,c,d)")
  print("E(a,b,c,d) = b(x).D(a,b,c,d) + d'<d>.B(a,b,c,d)")
  C.generate()
  print("S")

if __name__ == "__main__":
  main()