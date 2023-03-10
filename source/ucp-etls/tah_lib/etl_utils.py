import uuid
from datetime import datetime


def buildS3SubFolder():
    return datetime.now().strftime("%Y/%m/%d/%H/") + str(uuid.uuid1(node=None, clock_seq=None))
