#!/usr/bin/env python

import posixpath
import re

from re import Match
from mkdocs.config.defaults import MkDocsConfig
from mkdocs.structure.files import File, Files
from mkdocs.structure.pages import Page


def on_page_markdown(markdown: str, *, page: Page, config: MkDocsConfig, files: Files):
    """
    This hook is called after the page's markdown is loaded from file and can be
    """

    def replace(match: Match):
        badge_type, args = match.groups()
        args = args.strip()
        if badge_type == "version":
            return _badge_for_version(args, page, files)
        if badge_type == "scheduler":
            return _badge_for_scheduler(args, page, files)
        if type == "flag":
            return flag(args, page, files)

        # Otherwise, raise an error
        raise RuntimeError(f"Unknown shortcode: {badge_type}")

    return re.sub(r"<!-- md:(\w+)(.*?) -->", replace, markdown, flags=re.I | re.M)


def _badge(icon: str, text: str = "", badge_type: str = ""):
    """
    Create badge
    """
    classes = f"mdx-badge mdx-badge--{badge_type}" if badge_type else "mdx-badge"
    text = f"{text}{{ data-preview='' }}" if text.endswith(")") else text
    return "".join(
        [
            f'<span class="{classes}">',
            *([f'<span class="mdx-badge__icon">{icon}</span>'] if icon else []),
            *([f'<span class="mdx-badge__text">{text}</span>'] if text else []),
            "</span>",
        ]
    )


def _badge_for_scheduler(text: str, page: Page, files: Files):
    """
    Create badge for scheduler
    """

    # Return badge
    icon = "simple-docker"
    if text in ["kubernetes", "k3s"]:
        icon = "simple-kubernetes"
        text = "k3s"
    elif text == "nomad":
        icon = "simple-nomad"
    elif text in ["docker", "docker-local"]:
        icon = "simple-docker"
        text = "docker-local"

    href = f"https://dokku.com/docs/deployment/schedulers/{text}/"
    return _badge(
        icon=f"[:{icon}:]({href} '{text}')",
        text=f"[{text}]({href})",
    )


def _badge_for_version(text: str, page: Page, files: Files):
    """
    Create badge for version
    """

    # Return badge
    icon = "material-tag-outline"
    href = f"https://github.com/dokku/dokku/releases/tag/v{text}"
    return _badge(
        icon=f"[:{icon}:]({href} 'Since {text}')",
        text=f"[{text}]({href})",
    )


def flag(args: str, page: Page, files: Files):
    """
    Create flag
    """

    flag_type, *_ = args.split(" ", 1)
    if flag_type == "experimental":
        return _badge_for_experimental(page, files)
    raise RuntimeError(f"Unknown flag type: {flag_type}")


def _badge_for_experimental(page: Page, files: Files):
    """
    Create badge for experimental
    """

    icon = "material-flask-outline"
    href = _resolve_path("conventions.md#experimental", page, files)
    return _badge(icon=f"[:{icon}:]({href} 'Experimental')")


def _resolve_path(path: str, page: Page, files: Files):
    """
    Resolve path of file relative to given page - the posixpath always includes
    one additional level of `..` which we need to remove
    """

    path, anchor, *_ = f"{path}#".split("#")
    path = _resolve(files.get_file_from_path(path), page)
    return "#".join([path, anchor]) if anchor else path


def _resolve(file: File, page: Page):
    """
    Resolve path of file relative to given page - the posixpath always includes
    one additional level of `..` which we need to remove
    """

    path = posixpath.relpath(file.src_uri, page.file.src_uri)
    return posixpath.sep.join(path.split(posixpath.sep)[1:])
