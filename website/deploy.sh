#!/bin/bash

# Build the website
echo "Building website..."
npm run build

# Deploy to GitHub Pages
echo "Deploying to GitHub Pages..."
cd out
git init
git add .
git commit -m "Deploy website - $(date)"
git branch -m gh-pages
git remote add origin https://github.com/parzi-val/axle-file-sync.git
git push -f origin gh-pages
cd ..

echo "✅ Website deployed! Visit https://parzi-val.github.io/axle-file-sync/"