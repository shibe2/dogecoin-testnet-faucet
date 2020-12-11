# Imports

import os

# Get commit hash

commitHash = os.popen("git log --pretty=format:\"%h\" -1").read()

# Open the file to be added to and open output file

inFile = open("dist/indexVanilla.html", "r")
outFile = open("dist/index.html", "w")

# In future make file names command-line arguments as well as the replace values so we can also 
# replace the API_URL in frontend/src/mainUnbundled.js with this tool

lines = [] # All lines added to this

for line in inFile: # Loop through lines and add to lines
    lines.append(line)

lineCount = 0
found = False

# Loop through every line and replace the first instance of xxxxxxxx with the commit hash

while lineCount < len(lines) and found == False:
    if "xxxxxxxx" not in lines[lineCount]:
        lineCount += 1
    else:
        lines[lineCount] = lines[lineCount].replace("xxxxxxxx", "Commit Hash: " + commitHash)
        print(lines[lineCount])
        found = True

# Write new file to output

for line in lines:
    outFile.write(line)

# Close files

inFile.close()
outFile.close()

