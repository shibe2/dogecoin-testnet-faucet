commitHash = "testing1"

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
        found = True

for index in range(0, len(lines[lineCount])):
    if lines[lineCount][index] == "x":
        lineTest = lines[lineCount][index:index + 8]
        
        if lineTest.count("x") == len(lineTest):
            lines[lineCount] = lines[lineCount].replace("xxxxxxxx", commitHash)

            print(lines[lineCount])
            break

for line in lines:
    outFile.write(line)

inFile.close()
outFile.close()

