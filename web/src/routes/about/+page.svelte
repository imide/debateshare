<script lang="ts">
    import type { PageProps } from './$types';

    let { data }: PageProps = $props();
</script>

<section class="about">
    <h1>about</h1>
    <p>
        this project was created after some frustrations with speechdrop during tournaments. for example, speechdrop would occasionally fail to load, drop uploaded files and 
        <a href="https://github.com/yunyu/SpeechDrop/issues/46" target="_blank">not work sometimes.</a> in addition, speechdrop's codebase is
        <a href="https://github.com/yunyu/SpeechDrop/tree/master/src/main/java/edu/vanderbilt/yunyulin/speechdrop" target="_blank">quite dated (2017)</a>, thus potentially vulnerable to exploits, and has been 
        <a href="https://github.com/yunyu/SpeechDrop/issues/30" target="_blank">previously domain hijacked.</a> 
        <br>
        dont get me wrong, speechdrop is still wonderful and helped myself and plenty others during rounds, however 
        i wanted to negate some of the issues previously mentioned without burdening the original developer (<a href="https://github.com/yunyu" target="_blank">yunyu</a>) who seems to be out of college and thus no interest in debate.
        <br>
        plus i wanted to add some niceties, such as <strong>real-time emailing, post-round summaries,</strong> and other suggestions from 
        <a href="https://www.reddit.com/r/policydebate/comments/1fe10qk/speechdrop" target="_blank">reddit</a> and <strong>you! (hopefully)</strong>
    </p>
    <br>
    <h1>tech stack used</h1>
    <p>
        i used a mixture of tools and language to develop debateshare, such as:
    </p>
    <ul>
        <li>
            <a href="https://github.com/sveltejs/kit" target="_blank">sveltekit</a> + <a href="https://kendell.dev/m3-svelte/" target="_blank">m3-svelte</a> for the frontend
        </li>
        <li>
            golang + <a href="https://gin-gonic.com" target="_blank">gin</a> + <a href="https://github.com/go-gorm/gorm" target="_blank">gorm</a> + postgres for the backend
        </li>
    </ul>
    <p>all wrapped into a docker container hosted on (TODO: replace with actual host and add email too) </p>
    <br>
    <h1>contact</h1>
    <p>
        you can contact me at my email <strong>imide@imide.xyz</strong> or on the <a href="https://github.com/imide/debateshare">github repository.</a>
        <br>
        i am open to <strong>all and any suggestions you may have!</strong>
        <br>
        <br>
        <strong>this project cannot survive without donations.</strong>
        <br>
        if you like this project and have an extra cent to send, please checkout the README on the <a href="https://github.com/imide/debateshare">github repository.</a>
        <br>
        all extra donations past the required yearly cost will go to charity. the current charity is <strong><a href="https://www.pcrf.net/" target="_blank">palestine's children relief fund</a></strong> to aid those suffering in the gaza genocide and the greater israeli apartheid.
    </p>
</section>

<style>
    .about {
        width: min(100%, 56rem);
        margin-top: 1rem;
    }

    .about h1 {
        margin: 0 0 0.75rem;
    }

    .about li {
        margin: 0;
        color: var(--m3c-on-surface-variant);
        max-width: 70ch;
    }

    .about p {
        margin: 0;
        color: var(--m3c-on-surface-variant);
        max-width: 70ch;
    }
</style>