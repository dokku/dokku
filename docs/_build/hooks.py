#!/usr/bin/env python

import posixpath
import re

from collections.abc import Callable
from re import Match
from mkdocs.config.defaults import MkDocsConfig
from mkdocs.structure.files import File, Files
from mkdocs.structure.pages import Page


def on_page_markdown(markdown: str, *, page: Page, config: MkDocsConfig, files: Files) -> str | None:
    resp = re_on_page_markdown(markdown, page=page, config=config, files=files)
    if not resp:
        return resp
    

    modifiers: list[Callable[[list[str], str], tuple[list[str], bool]]] = [
        modify_content_noop,
        modify_content_links,
        modify_content_stripspace,
        modify_content_admonition,
        modify_content_terminal_example,
    ]

    is_modified = False
    lines = resp.split("\n")
    for modifier in modifiers:
        lines, modified = modifier(lines, page.file.src_path)
        if modified:
            is_modified = True

    if is_modified:
        resp = "\n".join(lines)
    
    return resp


def modify_content_noop(lines: list[str], _) -> tuple[list[str], bool]:
    """
    Simply returns the lines as is
    """
    modified = False
    updated_lines = []
    for line in lines:
        updated_lines.append(line)
    return updated_lines, modified


def modify_content_links(lines: list[str], filename: str) -> tuple[list[str], bool]:
    """
    Modifies links to be relative instead of absolute
    """
    filename = filename.replace("/usr/src/source/docs/", "")
    parts = filename.split("/")
    parts.pop()
    replacement = "](" + "/".join([".." for _ in parts]) + "/"
    modified = False
    updated_lines = []
    for line in lines:
        if "](/docs/" in line:
            line = line.replace("](/docs/", replacement)
            modified = True

        updated_lines.append(line)
    return updated_lines, modified


def modify_content_stripspace(lines: list[str], _) -> tuple[list[str], bool]:
    """
    Removes trailing whitespace from each line
    """
    modified = False
    updated_lines = []
    for line in lines:
        line = line.rstrip()
        updated_lines.append(line)
    return updated_lines, modified


def is_github_new(line: str, next_line: str | None) -> bool:
    """
    Checks if a given line is a github "new as of" admonition
    """
    if not next_line:
        return False

    return line.startswith("> [!IMPORTANT]") and "new as of" in next_line.lower()


def is_github_note(line: str) -> bool:
    """
    Checks if a given line is a github "note" admonition
    """
    return line.startswith("> [!NOTE]")


def is_github_warning(line: str) -> bool:
    """
    Checks if a given line is a github "warning" admonition
    """
    return line.startswith("> [!WARNING]")


def is_info(line: str) -> bool:
    """
    Checks if a given line is an "info" admonition
    """
    return line.startswith("> ")


def is_new(line: str) -> bool:
    """
    Checks if a given line is a "new as of" admonition
    """
    return line.startswith("> ") and "new as of" in line.lower()


def is_note(line: str) -> bool:
    """
    Checks if a given line is a "note" admonition
    """
    return line.startswith("> Note:")


def is_warning(line: str) -> bool:
    """
    Checks if a given line is a "warning" admonition
    """
    return line.startswith("> Warning:")


def modify_content_admonition(lines: list[str], filename: str) -> tuple[list[str], bool]:
    """
    Applies adminition info to each line in the output
    """
    modified = False
    updated_lines = []
    admonition_lines = []
    is_admonition = False
    replace_new_in_next_line = False
    for index, line in enumerate(lines):
        next_line: str | None = None
        if index + 1 < len(lines):
            next_line = lines[index + 1]

        if replace_new_in_next_line:
            line = line.replace("New as of", "Introduced in")
            line = line.replace("new as of", "introduced in")
            replace_new_in_next_line = False

        if is_github_new(line, next_line):
            line = line.replace("> [!IMPORTANT]", '!!! tip "New"')
            is_admonition = True
            admonition_lines.append(line)
            admonition_lines.append("")
            replace_new_in_next_line = True
        elif is_github_note(line):
            line = line.replace("> [!NOTE]", '!!! note "Note"')
            is_admonition = True
            admonition_lines.append(line)
            admonition_lines.append("")
        elif is_github_warning(line):
            line = line.replace("> [!WARNING]", '!!! warning "Warning"')
            is_admonition = True
            admonition_lines.append(line)
            admonition_lines.append("")
        elif is_new(line):
            print("is_new")
            line = line.replace("> ", '!!! tip "New"\n\n    ')
            line = line.replace("New as of", "Introduced in")
            line = line.replace("new as of", "introduced in")
            is_admonition = True
            admonition_lines.append(line)
        elif is_note(line):
            print("is_note")
            line = line.replace("> Note: ", '!!! note "Note"\n\n    ')
            is_admonition = True
            admonition_lines.append(line)
        elif is_warning(line):
            print("is_warning")
            line = line.replace("> Warning: ", '!!! warning "Warning"\n\n    ')
            is_admonition = True
            admonition_lines.append(line)
        elif is_info(line):
            if not is_admonition:
                line = line.replace("> ", '!!! info "Info"\n\n    ')
            elif line in [">", "> "]:
                line = ""
            else:
                line = "    " + line.removeprefix("> ")
            is_admonition = True
            admonition_lines.append(line)
        elif is_admonition and line in [">", "> "]:
            line = ""
            admonition_lines.append(line)
        elif is_admonition and line.startswith("> "):
            line = "    " + line.removeprefix("> ")
            admonition_lines.append(line)
        elif line == "":
            is_admonition = False
            if len(admonition_lines) > 0:
                modified = True
                updated_lines.extend(admonition_lines)
                admonition_lines = []
            updated_lines.append("")
        else:
            updated_lines.append(line)
    return updated_lines, modified


def is_shell_codeblock_start(line: str) -> bool:
    """
    Checks to see if a line starts a codeblock
    """
    return line == "```shell"


def modify_content_terminal_example(lines: list[str], _) -> tuple[list[str], bool]:
    """
    Modifies content so that terminal output is shown appropriately
    """
    modified = False
    updated_lines = []
    command_block = []
    example_block = []
    in_command_block = False
    in_example_block = False
    previous_block = ""
    next_line_must_be = None
    for line in lines:
        if is_shell_codeblock_start(line):
            command_block.append(line)
            modified = True
            in_command_block = True
            continue
        elif in_command_block:
            command_block.append(line)
            if line == "```":
                in_command_block = False
                previous_block = "command_block"
                next_line_must_be = ""
            continue
        elif line == "```":
            if previous_block == "":
                updated_lines.append(line)
                continue
            if previous_block == "command_block":
                example_block.append(line)

                if in_example_block:
                    previous_block = ""
                    in_example_block = False
                    updated_lines.append('=== "Shell"')
                    updated_lines.append("")
                    for command_line in command_block:
                        updated_lines.append(f"    {command_line}")
                    command_block = []

                    updated_lines.append("")
                    updated_lines.append('=== "Output"')
                    updated_lines.append("")
                    for example_line in example_block:
                        updated_lines.append(f"    {example_line}")
                    example_block = []
                else:
                    in_example_block = True
                continue
        elif previous_block == "command_block":
            if next_line_must_be is None:
                if in_example_block:
                    example_block.append(line)
                else:
                    updated_lines.extend(command_block)
                    updated_lines.append("")
                    updated_lines.append(line)
                    command_block = []
                    previous_block = ""
                continue
            if next_line_must_be == "":
                if line == "":
                    next_line_must_be = None
                    continue

        updated_lines.append(line)

    if len(command_block) > 0:
        updated_lines.extend(command_block)

    return updated_lines, modified


def re_on_page_markdown(markdown: str, *, page: Page, config: MkDocsConfig, files: Files) -> str | None:
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
