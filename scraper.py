import git
from bs4 import BeautifulSoup
import pathlib
import re
import json
from aiohttp import ClientSession
import asyncio
import msgpack
import pandas as pd
from tqdm import tqdm
import csv
from datetime import datetime
from shutil import rmtree
import os

CURRENT_FOLDER = os.path.dirname(os.path.abspath(__file__))

# Load dependency cache from file
try:
    with open("dependency_cache.msgpack", "rb") as infile:
        byte_data = infile.read()
    dependency_cache = msgpack.unpackb(byte_data)
except:
    dependency_cache = {}


# Fetch dependecy from npms
# TODO: does this scale? if we do millions of requests here will we be blocked?
async def fetch(session, item):
    url = "https://api.npms.io/v2/{endpoint}/{coin_name}".format(
        endpoint="package",
        coin_name=item
    )

    response = await session.get(url)
    json = await response.json()
    return json


async def get_repo_packages(repos):
    async with ClientSession() as session:
        for _, row in tqdm(repos.iterrows(), total=repos.shape[0], desc="fetching package.json files for repos"):
            # visit the main repo to find the file_finder_url (https://github.com/facebook/react/find/{master/main})
            repo_url = row["url"]
            repo_owner = row["url"].split("/")[-2]
            repo_name = row["url"].split("/")[-1]
            raw_repo_url = repo_url.replace(
                "https://github.com", "https://raw.githubusercontent.com")

            github_repo_page = await session.get(repo_url)
            repo_html = await github_repo_page.text()
            repo_html_bs_object = BeautifulSoup(
                repo_html, features="html.parser")

            # select the anchor tag that contains the file finder URL
            file_finder_anchor_tag = repo_html_bs_object.find(
                "a", href=re.compile("/find/"))
            if not hasattr(file_finder_anchor_tag, "attrs"):
                continue

            # visit the file finder page to find the URL to the filetree (it contains a weird hash at the end that we need)
            repo_file_finder_url = "https://github.com" + \
                file_finder_anchor_tag.attrs["href"]
            # get the main branch name while we're at it (https://github.com/facebook/react/find/master) --> master
            main_branch_name = repo_file_finder_url.split("/")[-1]
            repo_file_finder_page = await session.get(repo_file_finder_url)
            repo_file_finder_html = await repo_file_finder_page.text()
            repo_file_finder_bs_object = BeautifulSoup(
                repo_file_finder_html, features="html.parser")

            # select the tag that has the filetree URL
            repo_filetree_tag = repo_file_finder_bs_object.find("fuzzy-list")
            if not hasattr(repo_filetree_tag, "attrs"):
                continue

            # visit the filetree with the requested-with header (it 400's without it, maybe anti-scrape measure?)
            repo_filetree_url = "https://github.com" + \
                repo_filetree_tag.attrs["data-url"]
            session.headers.update({"X-Requested-With": "XMLHttpRequest"})
            repo_filetree_page = await session.get(repo_filetree_url)
            repo_filetree_json = await repo_filetree_page.json()

            # the files in the repo are under the json key 'paths'
            repo_paths = repo_filetree_json["paths"]

            # find and collect urls to packagefiles in filetree
            package_file_paths = [
                raw_repo_url + "/" + main_branch_name + "/" + path for path in repo_paths if "package.json" in path and "node_modules" not in path]

            print(repo_filetree_url, package_file_paths)

            for idx, package_file_url in enumerate(package_file_paths):
                package_file_page = await session.get(package_file_url)
                package_file = await package_file_page.text()
                filename = os.path.join(
                    CURRENT_FOLDER, "repos", repo_owner + "_" + repo_name, "package" + str(idx) + ".json")
                # create dir if it doesn't exist and write package.json files
                os.makedirs(os.path.dirname(filename), exist_ok=True)
                with open(filename, "w+") as file:
                    file.write(package_file)


async def get_repo_dependencies(repo_url, repo_id):
    # Find all package.json files in repo
    paths = filter(lambda r: "node_modules" not in r, list(
        pathlib.Path("repos/" + repo_id).glob("**/package.json")))

    # Parse package.json
    dependency_collector = set()
    for path in paths:
        with open(str(path)) as file:
            try:
                package = json.load(file)
                dependency_collector.update(
                    list(package["devDependencies"].keys()))
                dependency_collector.update(
                    list(package["devDependencies"].keys()))
            except:
                continue

    # Replace @ and / with hexadecimal counterparts
    dependencies = [dependency.replace(
        "@", "%40").replace("/", "%2f") for dependency in dependency_collector]

    urls = []

    async with ClientSession() as session:
        for dependency in dependencies:

            if dependency in dependency_cache:
                urls.append(dependency_cache[dependency])

            else:
                response = await fetch(session, dependency)

                try:
                    url = response["collected"]["metadata"]["links"]["repository"]
                    urls.append(url)
                    dependency_cache[dependency] = url
                except Exception:
                    # TODO: we could also check if there is a link in collected/repository/url
                    # but sometimes they are ssh and sometimes http links so would need massageing
                    pass
    return urls


async def main():
    # load the data dump from bigquery
    repos = pd.read_csv("test.csv")

    # format links to point to actual github repo URL
    repos['url'] = repos['url'].str.replace(
        "api.github.com/repos", "github.com")

    await get_repo_packages(repos)

    res = set()

    # clone_repos(repos)

    last_backup = datetime.now().second

    for _, row in tqdm(repos.iterrows(), total=repos.shape[0], desc="fetching dependencies"):

        # Backup every 15 minutes
        if datetime.now().second - last_backup >= 900:
            # msgpack is a binary json format ;)
            with open("dependency_cache.msgpack", "wb") as outfile:
                packed = msgpack.packb(dependency_cache)
                outfile.write(packed)
            with open("res.csv", "w") as outfile:
                writer = csv.writer(outfile, lineterminator="\n")
                for tup in res:
                    writer.writerow(tup)
            last_backup = datetime.now().second

        urls = await get_repo_dependencies(row["url"], str(row["id"]))

        for url in urls:
            res.add((row["id"], url))

    # Write result to csv
    # TODO: Now we overide each time, do we want to compile all results in one csv or in smaller ones?
    with open("res.csv", "w") as outfile:
        writer = csv.writer(outfile, lineterminator="\n")
        for tup in res:
            writer.writerow(tup)

    # Save cache
    with open("dependency_cache.msgpack", "wb") as outfile:
        packed = msgpack.packb(dependency_cache)
        outfile.write(packed)

    # Remove the cloned repos
    rmtree("repos")


if __name__ == "__main__":
    asyncio.run(main())
