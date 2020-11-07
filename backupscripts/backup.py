"""
Uploads the file at the given path to dropbox with the given access token.
"""


import sys

import dropbox
from dropbox.exceptions import ApiError, AuthError
from dropbox.files import WriteMode

def uploadToDropbox(path, token):

    with dropbox.Dropbox(token) as dbx:
        try:
            # Check token is valid
            dbx.users_get_current_account()

            with open(path, 'rb') as f:
                try:
                    # Try to uplaod user's data
                    dbx.files_upload(f.read(), path, mode=WriteMode('overwrite'))

                except AuthError:
                    sys.exit("ERROR: Invalid access token.")

        except ApiError:
            sys.exit("ERROR: User may be out of space.")
