"""
Uploads the file at the given path to dropbox with the given access token.
"""


import sys

import dropbox
from dropbox.exceptions import ApiError, AuthError
from dropbox.files import WriteMode


def uploadToDropbox(path: str, token: str) -> None:

    with dropbox.Dropbox(token) as dbx:
        try:
            # Check token is valid
            dbx.users_get_current_account()

            with open(path, "rb") as f:
                try:
                    # Try to uplaod user's data
                    dbxPath = path
                    dbx.files_upload(f.read(), dbxPath, mode=WriteMode("overwrite"))

                except AuthError:
                    sys.exit("ERROR: Invalid access token.")

        except ApiError:
            sys.exit("ERROR: User may be out of space.")


if __name__ == "__main__":

    path = sys.argv[1]
    token = sys.argv[2]

    # print("Uploading user data to dropbox...")
    uploadToDropbox(path, token)
    # print("Done :)")
