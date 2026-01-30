(() => {
    const glitch = t => {
        for (var r=[[/o/g,'ө'],[/O/g,'Ө'],[/e/g,'ә'],[/E/g,'Ә'],[/Y/g,'Ұ'],[/H/g,'Ң']],c=[],m,i=0;i<r.length;i++)
            for (var g=new RegExp(r[i][0].source,'g'); (m=g.exec(t)); )
                c.push([m.index,r[i][1]]);
        return c.length
            ? (m=c[Math.random()*c.length|0], t.slice(0,m[0])+m[1]+t.slice(m[0]+1))
            : t;
    };

    document.addEventListener('DOMContentLoaded', () => {
        document.querySelectorAll('.bot-gl').forEach(el => {
            const clean = el.innerHTML;

            el.innerHTML = clean.replace(/[^<>]+/g, s => glitch(s));
            el.addEventListener('mouseenter', () => {
                el.innerHTML = clean;
            });

            el.addEventListener('mouseleave', () => {
                el.innerHTML = clean.replace(/[^<>]+/g, s => glitch(s));
            });
        });
    });
})();