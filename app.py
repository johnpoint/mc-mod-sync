import os
import hashlib
import json
import shutil
import sys
import zipfile
import requests


def cmd():
    if len(sys.argv) < 2:
        sys.exit(0)
    if sys.argv[1] == "gen":
        gen_zip()
    if sys.argv[1] == "get" and len(sys.argv) == 3:
        get_mod()
    print("All done! Have fun!")


def get_mod():
    files = os.listdir(".")
    now_mod_map = {}
    need_enable = []
    need_download = []
    need_disable = {}

    for i in files:
        if ".jar" not in i:
            continue
        try:
            with open(i, "rb") as fp:
                data = fp.read()
            file_md5 = hashlib.md5(data).hexdigest()
            now_mod_map[file_md5] = i
            need_disable[file_md5] = i
        except:
            pass

    print("! Featch config.json from " + sys.argv[2])
    update_config = requests.get(sys.argv[2]).text
    update = json.loads(update_config)

    for i in update:
        if now_mod_map.__contains__(i["md5"]):
            print(
                "\033[0;32m+",
                i["md5"][0:7],
                i["name"] + "\033[0m",
            )
            os.rename(now_mod_map[i["md5"]], i["name"])
            need_disable.pop(i["md5"])
        else:
            print("\033[0;33m↓", i["md5"][0:7], i["name"] + "\033[0m")
            download(sys.argv[2].replace("update.json", "") + i["md5"], i["name"])
    for k in need_disable:
        if ".disable" in need_disable[k]:
            continue
        print(
            "\033[0;36mx",
            k[0:7],
            need_disable[k].replace(".disable", "", -1) + "\033[0m",
        )
        os.rename(
            need_disable[k], need_disable[k].replace(".disable", "", -1) + ".disable"
        )
    pass


def download(url, name):
    r = requests.get(url).content
    with open(name, "wb") as f:
        f.write(r)


def gen_zip():
    files = os.listdir(".")
    mod_info = []
    zip_files = []
    name_map = {}

    try:
        shutil.rmtree(".mcmodtmp")
        os.remove("mod.zip")
    except:
        pass

    os.mkdir(".mcmodtmp")

    print("! Sanning")
    for i in files:
        if (".disable" in i) or (".jar" not in i):
            continue
        try:
            with open(i, "rb") as fp:
                data = fp.read()
            file_md5 = hashlib.md5(data).hexdigest()
            with open(".mcmodtmp/" + file_md5, "wb") as wf:
                wf.write(data)
            mod_info.append({"name": i, "md5": file_md5})
            zip_files.append(".mcmodtmp/" + file_md5)
            name_map[file_md5] = i
        except:
            pass

    with open(".mcmodtmp/update.json", "w") as f:
        json.dump(mod_info, f)
        zip_files.append(".mcmodtmp/update.json")
        name_map["update.json"] = "update.json"

    zip = zipfile.ZipFile("mod.zip", "w", zipfile.ZIP_DEFLATED)
    print("! Packing")
    for file in zip_files:
        print(
            "+",
            file.replace(".mcmodtmp/", "")[0:7],
            name_map[file.replace(".mcmodtmp/", "")],
        )
        zip.write(file, arcname=file.replace(".mcmodtmp/", ""))
    zip.close()

    try:
        shutil.rmtree(".mcmodtmp")
    except:
        pass


if __name__ == "__main__":
    cmd()
