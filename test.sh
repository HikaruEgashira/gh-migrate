LC_ALL=C find . -type f -not -path '*/.*' -exec \
    sed -i '' 's#migrate#transform#g' {} +
