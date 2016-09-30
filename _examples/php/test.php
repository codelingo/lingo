<?php

$global_var_1 = "hi";

// another comment
function mysql_test() {
	if (5 > 3) {
		echo "hi";
	}
	else {
		echo "bye";
	}
// asdf
	$link = mysql_connect('mysql_host', 'mysql_user', 'mysql_password');
	if (!$link) {
		echo 'Could not connect to mysql';
		exit;
	}
}

function say($a) {

	echo "hi" + $a;
}