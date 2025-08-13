# TODO
Be able to convert markdown files to html
upload that site to github
document work on some projects that im working on

have a folder of md files that represents the structure of website.
figure out how to translate md files to html and css.

figure out how styling / themes should work, maybe i can define a theme.css file 
    that file could define styling for all the elements that this ssg will provide
    like it could provide styles for components like the nav bar, blog list, fonts, etc
    have a themes/ folder with css files. the name of the css files can be references in the ssg.toml file
        theme = "gruvbox"
        theme = "rosepine"

this is mostly for dev logs / personal stuff
be able to add blogs in a content/ directory
those blogs get added to the /blog/{blog-name} end point

be able to add custom pages like /about
maybe each end point is just a directory?

the root / end point should just be a list of blogs
follow matklads stlying on this

there should be a home / nav bar with all the end points listed there so stuff like about, blogs, etc
should also have the author's name on the top left

there should also be a thing at the bottom with stuff like my github linked.
again follow matklad's website on this

be able to add an rss feed to the blog, this should be configurable.
    be able to configure how many blogs to include based on date, etc
    
blogs should have meta data like date published, description, is the blog a draft etc

figure out an easy way to deploy

development web server
    hot reloading

usage:
    to add a new blog just create a new file in content/
    to add images / other resources just add files to static/
    the markdown should be ergonomic and easy to work with
        - be able to easily add links, images, videos, code snippets
    easily add a description within the markdown (maybe it could be shown in
        the blog list as a preview)
    one command deployment to github pages
    automatically generated rss feed
    easily add end points that get added to the nav bar. stuff like about, etc
    create a .toml file in the root to specify basic stuff like:
        - deployment location
        - project name
        - theme
        - should drafts be built

### Tomorrow

    -- create html templates to integrate meta data into
    -- create an index.html file that contains a blog list
    -- fix server/. write a proper handler for routing to nodes
    -- figure out how to derive a node's name

    - toml config file. (themes, website name, deployment info)
        integrate this toml file into site/
        it should parse this file and create some struct using it
        the struct can then be used to alter the generated html. so stuff like themes can be applied on to it

    - get themes working - there should be a themes/ dir with css files that you can choose in the config file
    
    fix site tests

    -- read metadata from blogs
        - author
        - title
        - date published
        - description
        - tags [zig, go, programming, etc]
        - draft

    -- create a web server using site.Nodes
    -- in nodes the file extension should be replaced from ".md" to ".html" when converting
    be able to fully build the project, copy hugo on this, place html files where they place them

    figure out how to convert links in md files to other md files to html links

    if a directory is built using . or ./ then figure out how to properly strip paths
    test server/
    
    give the user access to site data using special commands like @title() / @author()
        or maybe go's templates can still be used?
    be able to place html files and markdown files. html files still contain the required metadata using +++
