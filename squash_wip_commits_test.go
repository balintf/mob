package main

import (
	"io/ioutil"
	"testing"
)

// TODO if last commit is wip commit, exit with warning
func TestSquashWipCommits_acceptance(t *testing.T) {
	_, configuration := localSetup(t)
	start(configuration)
	createFile(t, "file2.txt", "owqe")
	next(configuration)
	start(configuration)
	createFileAndCommitIt(t, "file1.txt", "owqe", "new file")

	squashWipCommits(configuration)

	equals(t, []string{
		"new file",
	}, commitsOnCurrentBranch(configuration))
}

func TestCommitsOnCurrentBranch(t *testing.T) {
	_, configuration := localSetup(t)
	createFileAndCommitIt(t, "file1.txt", "irrelevant", "not on branch")
	start(configuration)
	createFileAndCommitIt(t, "file2.txt", "irrelevant", "on branch")
	createFile(t, "file3.txt", "irrelevant")
	next(configuration)
	start(configuration)

	commits := commitsOnCurrentBranch(configuration)

	equals(t, []string{
		configuration.WipCommitMessage,
		"on branch",
	}, commits)
}

func TestEndsWithWipCommit_finalManualCommit(t *testing.T) {
	_, configuration := localSetup(t)
	start(configuration)
	createFileAndCommitIt(t, "file1.txt", "owqe", "new file")

	equals(t, false, endsWithWipCommit(configuration))
}

func TestEndsWithWipCommit_finalWipCommit(t *testing.T) {
	_, configuration := localSetup(t)
	start(configuration)
	createFile(t, "file1.txt", "owqe")
	next(configuration)
	start(configuration)

	equals(t, true, endsWithWipCommit(configuration))
}

func TestEndsWithWipCommit_manualThenWipCommit(t *testing.T) {
	_, configuration := localSetup(t)
	start(configuration)
	createFileAndCommitIt(t, "file1.txt", "owqe", "new file")
	createFile(t, "file2.txt", "owqe")
	next(configuration)
	start(configuration)

	equals(t, true, endsWithWipCommit(configuration))
}

func TestEndsWithWipCommit_wipThenManualCommit(t *testing.T) {
	_, configuration := localSetup(t)
	start(configuration)
	createFile(t, "file2.txt", "owqe")
	next(configuration)
	start(configuration)
	createFileAndCommitIt(t, "file1.txt", "owqe", "new file")

	equals(t, false, endsWithWipCommit(configuration))
}

func TestMarkSquashWip_singleManualCommit(t *testing.T) {
	configuration = getDefaultConfiguration()
	input := "pick c51a56d new file\n" +
		"\n" +
		"# Rebase ..."

	result := markPostWipCommitsForSquashing(input, configuration)

	equals(t, input, result)
}

func TestMarkSquashWip_manyManualCommits(t *testing.T) {
	configuration = getDefaultConfiguration()
	input := "pick c51a56d new file\n" +
		"pick 63ef7a4 another commit\n" +
		"\n" +
		"# Rebase ..."

	result := markPostWipCommitsForSquashing(input, configuration)

	equals(t, input, result)
}

func TestMarkSquashWip_wipCommitFollowedByManualCommit(t *testing.T) {
	configuration = getDefaultConfiguration()
	input := "pick 01a9a31 " + configuration.WipCommitMessage + "\n" +
		"pick c51a56d manual commit\n" +
		"\n" +
		"# Rebase ..."
	expected := "pick 01a9a31 " + configuration.WipCommitMessage + "\n" +
		"squash c51a56d manual commit\n" +
		"\n" +
		"# Rebase ..."

	result := markPostWipCommitsForSquashing(input, configuration)

	equals(t, expected, result)
}

func TestMarkSquashWip_manyWipCommitsFollowedByManualCommit(t *testing.T) {
	configuration = getDefaultConfiguration()
	input := "pick 01a9a31 " + configuration.WipCommitMessage + "\n" +
		"pick 01a9a32 " + configuration.WipCommitMessage + "\n" +
		"pick 01a9a33 " + configuration.WipCommitMessage + "\n" +
		"pick c51a56d manual commit\n" +
		"\n" +
		"# Rebase ..."
	expected := "pick 01a9a31 " + configuration.WipCommitMessage + "\n" +
		"squash 01a9a32 " + configuration.WipCommitMessage + "\n" +
		"squash 01a9a33 " + configuration.WipCommitMessage + "\n" +
		"squash c51a56d manual commit\n" +
		"\n" +
		"# Rebase ..."

	result := markPostWipCommitsForSquashing(input, configuration)

	equals(t, expected, result)
}

func TestCommentWipCommits_oneWipAndOneManualCommit(t *testing.T) {
	configuration = getDefaultConfiguration()
	input := "# This is a combination of 2 commits.\n" +
		"# This is the 1st commit message:\n" +
		"\n" +
		configuration.WipCommitMessage + "\n" +
		"\n" +
		"# This is the commit message #2:\n" +
		"\n" +
		"manual commit\n" +
		"\n" +
		"# Please enter ..."
	expected := "# This is a combination of 2 commits.\n" +
		"# This is the 1st commit message:\n" +
		"\n" +
		"# " + configuration.WipCommitMessage + "\n" +
		"\n" +
		"# This is the commit message #2:\n" +
		"\n" +
		"manual commit\n" +
		"\n" +
		"# Please enter ..."

	result := commentWipCommits(input, configuration)

	equals(t, expected, result)
}

func TestSquashWipCommitGitEditor(t *testing.T) {
	createTestbed(t)
	path := createFile(t, "commits",
		"# This is a combination of 2 commits.\n"+
			"# This is the 1st commit message:\n \n"+
			"mob next [ci-skip] [ci skip] [skip ci]\n \n"+
			"# This is the commit message #2:\n \n"+
			"new file\n \n"+
			"# Please enter the commit message for your changes. Lines starting\n")
	expected := "# This is a combination of 2 commits.\n" +
		"# This is the 1st commit message:\n \n" +
		"# mob next [ci-skip] [ci skip] [skip ci]\n \n" +
		"# This is the commit message #2:\n \n" +
		"new file\n \n" +
		"# Please enter the commit message for your changes. Lines starting\n"

	squashWipCommitsGitEditor(path, getDefaultConfiguration())

	result, _ := ioutil.ReadFile(path)
	equals(t, expected, string(result))
}

func TestSquashWipCommitGitSequenceEditor(t *testing.T) {
	createTestbed(t)
	configuration = getDefaultConfiguration()
	path := createFile(t, "rebase",
		"pick 01a9a31 "+configuration.WipCommitMessage+"\n"+
			"pick 01a9a32 "+configuration.WipCommitMessage+"\n"+
			"pick 01a9a33 "+configuration.WipCommitMessage+"\n"+
			"pick c51a56d manual commit\n"+
			"\n"+
			"# Rebase ...\n")
	expected := "pick 01a9a31 " + configuration.WipCommitMessage + "\n" +
		"squash 01a9a32 " + configuration.WipCommitMessage + "\n" +
		"squash 01a9a33 " + configuration.WipCommitMessage + "\n" +
		"squash c51a56d manual commit\n" +
		"\n" +
		"# Rebase ...\n"

	squashWipCommitsGitSequenceEditor(path, getDefaultConfiguration())

	result, _ := ioutil.ReadFile(path)
	equals(t, expected, string(result))
}
