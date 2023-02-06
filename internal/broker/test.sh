# shellcheck disable=SC2034
for (( i = 0 ; i <= 50; i++ ))
do
  echo "-----_____run test_______" $i
  go test
done
