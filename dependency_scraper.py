import time
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC

chrome_options = Options()
# chrome_options.add_argument("--headless")

driver = webdriver.Chrome(options=chrome_options)
driver.get("https://github.com/facebook/react/network/dependencies")


# press all pagination buttons to expand all dependencies
while True:
    try:
        time.sleep(2)
        load_more_btn = driver.find_element_by_css_selector(
            "button.ajax-pagination-btn")
        load_more_btn.click()
    except:
        break

# process dependencies
processed = set()
dependency_lists = driver.find_elements_by_css_selector("div.Box.mb-3")

for dependency_list in dependency_lists:
    header = dependency_list.find_element_by_css_selector(".Box-header")
    dependencies_source = header.find_element_by_tag_name("a").text
    if "yarn.lock" in dependencies_source or "package-lock.json" in dependencies_source:
        continue
    dependencies = dependency_list.find_elements_by_css_selector(
        ".Box-row.d-flex.flex-items-center")
    for dependency in dependencies:
        repository_url = dependency.find_element_by_css_selector(
            "a[data-octo-click=\"dep_graph_package\"]").get_attribute("href")
        # print(repository_url)
        processed.add(repository_url)


print(processed)
print(len(processed))
