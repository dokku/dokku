import os

environment = os.getenv('ENVIRONMENT')

if environment == 'dev':
    from .dev import *
elif environment == 'staging':
    from .staging import *
elif environment == 'production':
    from .staging import *
else:
    raise Exception("Invalid environment")
