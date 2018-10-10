# SLEM (aka. "Shaking Dog") Register

This is a Golang API webserver project I built based off some requirements given by a friend. It meets those requirements, but won't really make sense to anyone else. It is really only public to form part of my portfolio, so feel free to take a look and provide feedback. It was my second venture into Golang, and what an amazing language it is. Except for the lack of generics (tsk tsk).

A **few** random things to note:

* it lacks tests (naughty me I know)
* it's built for a single server deploy (but could scale if auth cookie secret is shared between instances, and data synchronisation is throughly tested)
* it uses [Gorilla Web Toolkit](http://www.gorillatoolkit.org/pkg/) libraries
* backs onto a MySQL DB
* requires an [Okta](https://www.okta.com/) or similar OAuth2 OpenID provider

An implementation of a UI can be found [here](https://github.com/ishkanan/shakingdog-ui).
