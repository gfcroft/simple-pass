#!/usr/bin/env bash

# NOTE: consider setting pipefail if we start using pipes
# set -uxo pipefail 
set -ux #print out commands to term & err if we use undeclared vars, with the exceptions of *and@
# set -u #err if we use undeclared vars, with the exceptions of *and@

RED='\033[0;31m'
GREEN='\033[0;32m'
ORANGE='\033[0;33m'
NC='\033[0m' # No Color


must_succeed(){
    ret=$1
    if [ "$ret" -ne 0 ]; then
        echo -e "${RED}TEST FAILED - expected zero return from last cmd executed${NC}"
        exit 1
    fi
}

print_test_pass(){
    echo -e "${GREEN}TEST PASSED${NC}"
}

must_fail(){
    ret=$1
    if [ "$ret" -eq 0 ]; then
        echo -e "${RED}TEST FAILED - expected non-zero return from last cmd executed${NC}"
        exit 1
    fi
}

test_case_print(){
    echo -e "###############${ORANGE}$*${NC}"
}

######### MAIN #######

# TODO someway of getting this cleanly from internal/commmon/constants
if [ -z "${PASSDB_DEV-}" ]; then
    export PASSDB_DEV="True"
fi

TEST_CACHE_FILE=~/.passdb.dev

# avoid issues with old dev cache files
if [ -f $TEST_CACHE_FILE ]; then
    rm $TEST_CACHE_FILE
fi


SIMPLE_PASS_CMD=bin/simple-pass

VALID_PASSWORD="Password^123"
INVALID_PASSWORD="0"

VALID_PASSDB_NAME="testDB"
INVALID_PASSDB_NAME=""

VALID_PATH="/tmp/${RANDOM}"
TEMP_TEST_DIR=$(mktemp -d)


VALID_ITEM_NAME="some-name"
ALT_VALID_ITEM_NAME="alt-${VALID_ITEM_NAME}"
VALID_ITEM_USERNAME="john.smith"
VALID_ITEM_PASSWORD="foo"
VALID_ITEM_URL="some-name.eg.com"
VALID_ITEM_NOTES="blah blah"

#simple-pass == sass
test_case_print "should print help and fail when no command given"
$SIMPLE_PASS_CMD | grep "the help"
must_fail $(( "${PIPESTATUS[@]/%/ +} 0"))
print_test_pass

# TODO test the text output of cmds in a useful, but not fragile way? (e.g. match part of the text, not all) 
# TODO failure tests really need to grep the output of error messages in order to be sure that they are failing for the right reasons
test_case_print "should fail because no name, path or password"
$SIMPLE_PASS_CMD create-pass-db # should fail because no name or path
must_fail $?
print_test_pass

test_case_print "should fail because no name"
$SIMPLE_PASS_CMD create-pass-db --filePath $VALID_PATH --password $VALID_PASSWORD
must_fail $?
print_test_pass

test_case_print "should fail because no path"
$SIMPLE_PASS_CMD create-pass-db --name $VALID_PASSDB_NAME --password $VALID_PASSWORD  # TODO have a default path?
must_fail $?
print_test_pass


test_case_print "should fail because invalid path"
existing_file_path=/tmp/existing-pass-db-test
touch $existing_file_path
$SIMPLE_PASS_CMD create-pass-db --name $VALID_PASSDB_NAME --filePath $existing_file_path --password $VALID_PASSWORD
must_fail $?
rm $existing_file_path
print_test_pass

test_case_print "should fail because invalid password"
$SIMPLE_PASS_CMD create-pass-db --name $VALID_PASSDB_NAME --filePath $existing_file_path --password $INVALID_PASSWORD
must_fail $?
rm $existing_file_path
print_test_pass


test_case_print "should successfully create pass db"
test_pass_db_path="${TEMP_TEST_DIR}/test-a"
$SIMPLE_PASS_CMD create-pass-db --name $VALID_PASSDB_NAME --filePath "$test_pass_db_path" --password $VALID_PASSWORD
must_succeed $?
print_test_pass


test_case_print "should load pass db (using cache) and return info"
info_test_output_file=/tmp/test-pass-db-info-test.tmp
$SIMPLE_PASS_CMD status | tee $info_test_output_file
# must_succeed $(( "${PIPESTATUS[@]/%/ +} 0"))
must_succeed $(( "${PIPESTATUS[@]/%/ +} 0"))
if  ! grep "$VALID_PASSDB_NAME" < "$info_test_output_file" > /dev/null ; then
    echo "failed to retrieve correct info - did not see expected passdb name ('${VALID_PASSDB_NAME}') in output"
    exit 1
fi
print_test_pass

test_case_print "should load pass db (without using cache)"
test_pass_db_path="${TEMP_TEST_DIR}/test-a"
[ ! -f $TEST_CACHE_FILE ] && echo "test cache file: $TEST_CACHE_FILE does not exist!" && exit 1
rm $TEST_CACHE_FILE
$SIMPLE_PASS_CMD load-pass-db --filePath "$test_pass_db_path" --password $VALID_PASSWORD
must_succeed $?
print_test_pass


test_case_print "should list 0 passdb items successfully"
info_test_output_file=/tmp/test-pass-db-info-test.tmp
$SIMPLE_PASS_CMD list
must_succeed $?
if  ! grep "$VALID_PASSDB_NAME" < "$info_test_output_file" > /dev/null ; then
    echo "failed to retrieve correct info - did not see expected passdb name ('${VALID_PASSDB_NAME}') in output"
    exit 1
fi
print_test_pass


# TODO improve this
cat <<EOF
##
# WARNING - THE FOLLOWING TESTS ARE ALL LINKED TO THE SAME DATASTORE
# TODO improve
##
EOF

test_case_print "should add valid new passdb item"
$SIMPLE_PASS_CMD add "$VALID_ITEM_NAME" --username "$VALID_ITEM_USERNAME" --password "$VALID_ITEM_PASSWORD" --url "$VALID_ITEM_URL" --notes "$VALID_ITEM_NOTES"
must_succeed $?
print_test_pass


test_case_print "should fail to add invalid new passdb item"
$SIMPLE_PASS_CMD add "" #name can't be blank, rest can
must_fail $?
print_test_pass


test_case_print "should retrieve valid passdb item" #TODO grep this 
$SIMPLE_PASS_CMD get "$VALID_ITEM_NAME"
must_succeed $?
print_test_pass


test_case_print "should retrieve valid passdb item part"
declare -A flag_val_map
flag_val_map["username"]="$VALID_ITEM_USERNAME"
flag_val_map["u"]="$VALID_ITEM_USERNAME"

flag_val_map["url"]="$VALID_ITEM_URL"
flag_val_map["w"]="$VALID_ITEM_URL"

flag_val_map["password"]="$VALID_ITEM_PASSWORD"
flag_val_map["p"]="$VALID_ITEM_PASSWORD"

flag_val_map["notes"]="$VALID_ITEM_NOTES"
flag_val_map["n"]="$VALID_ITEM_NOTES"

for flag in "username" "u" "url" "w" "password" "p" "notes" "n"; do 
    expected_val="${flag_val_map[$flag]}" 
    if (( ${#flag} < 2 )); then 
        #must be a shorthand flag
        $SIMPLE_PASS_CMD get $VALID_ITEM_NAME -"${flag}" | tee ${flag}.tmp
    else
        #must be a full length flag
        $SIMPLE_PASS_CMD get $VALID_ITEM_NAME --"${flag}" | tee ${flag}.tmp
    fi

    must_succeed $(( "${PIPESTATUS[@]/%/ +} 0"))

    if ! grep "$expected_val" < "${flag}.tmp" > /dev/null ; then
        echo "did not retrieve expected value $expected_val for flag of item: $flag"
        echo "got: $(cat ${flag}.tmp)"
        rm ${flag}.tmp
        exit 1
    fi
    rm ${flag}.tmp
done
print_test_pass



test_case_print "should fail to retrieve non-existant passdb item"
$SIMPLE_PASS_CMD get $RANDOM
must_fail $?
print_test_pass

$SIMPLE_PASS_CMD get $RANDOM --url
must_fail $?
print_test_pass

test_case_print "should rename valid passdb item"
$SIMPLE_PASS_CMD rename $VALID_ITEM_NAME --to $ALT_VALID_ITEM_NAME
must_succeed $?

$SIMPLE_PASS_CMD get $ALT_VALID_ITEM_NAME 
must_succeed $?

# now put it back
$SIMPLE_PASS_CMD rename $ALT_VALID_ITEM_NAME --to $VALID_ITEM_NAME
must_succeed $?

print_test_pass




test_case_print "should not rename invalid passdb item"
$SIMPLE_PASS_CMD rename $VALID_ITEM_NAME --to $VALID_ITEM_NAME
must_fail $?

$SIMPLE_PASS_CMD rename $VALID_ITEM_NAME --to "" 
must_fail $?

print_test_pass




test_case_print "should update valid passdb item"
new_note="new-note"
# TODO do for all parts?
$SIMPLE_PASS_CMD update $VALID_ITEM_NAME --notes "$new_note"
must_succeed $?

retrieved_note=$($SIMPLE_PASS_CMD get $VALID_ITEM_NAME --notes )
if ! echo "$retrieved_note" | grep "new-note" > /dev/null; then
    echo "failed to update item '$VALID_ITEM_NAME' - new note: '$new_note' does not match expected '$retrieved_note'"
    exit 1
fi
print_test_pass

test_case_print "should fail to update invalid passdb item"
$SIMPLE_PASS_CMD update $RANDOM
must_fail $?
print_test_pass

test_case_print "should delete valid passdb item"
$SIMPLE_PASS_CMD delete $VALID_ITEM_NAME
must_succeed $?


if $SIMPLE_PASS_CMD list | grep $VALID_ITEM_NAME > /dev/null ; then
    echo "failed to delete item $VALID_ITEM_NAME - still listed in output of list:"
    $SIMPLE_PASS_CMD list
    exit 1
fi
print_test_pass


test_case_print "should fail to delete invalid passdb item"
$SIMPLE_PASS_CMD delete $RANDOM
must_fail $?
print_test_pass


