import os

commitHash = os.popen("git log --pretty=format:\"%h\" -1").read()

inFile = open("../src/indexVanilla.html", "r")
outFile = open("../src/index.html", "w")

lines = []

for line in inFile:
    lines.append(line)

lineCount = 0
found = False

while lineCount < len(lines) and found == False:
    if "xxxxxxxx" not in lines[lineCount]:
        lineCount += 1
    else:
        lines[lineCount] = lines[lineCount].replace("xxxxxxxx", "Commit Hash: " + commitHash)
        print(lines[lineCount])
        found = True

for line in lines:
    outFile.write(line)

inFile.close()
outFile.close()

