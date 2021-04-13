#!/bin/bash
i=0
for FILE in in/*; do
  echo "==================================="
  echo "| working on file: $FILE |"
  echo "==================================="
  echo ""

  ./KEX $FILE "out/res$i.csv"
  ((i++))
  echo "==================================="
  echo "| Done with file: $FILE |"
  echo "==================================="
  echo ""
done
