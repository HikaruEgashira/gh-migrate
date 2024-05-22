LC_ALL=C find . -type f ! -name ".*" -exec \
    sed -i '' 's#transform#transform#g' {} +
