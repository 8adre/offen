from os import environ

from flask import Flask
from flask_sqlalchemy import SQLAlchemy
from flask_admin import Admin

app = Flask(__name__)
app.secret_key = environ.get("SESSION_SECRET")
app.config["SQLALCHEMY_DATABASE_URI"] = environ.get("MYSQL_CONNECTION_STRING")
db = SQLAlchemy(app)

from accounts.models import Account, User
from accounts.views import AccountView, UserView
import accounts.api

app.config["FLASK_ADMIN_SWATCH"] = "flatly"

admin = Admin(app, name="offen admin", template_mode="bootstrap3", base_template="index.html")

admin.add_view(AccountView(Account, db.session))
admin.add_view(UserView(User, db.session))