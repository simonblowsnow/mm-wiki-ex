cd mm-wiki-ex
git fetch gitee
git reset --hard gitee/master
cd ..
rm -rf mm-wiki-ex/conf/mm-wiki.conf
cp conf/mm-wiki.conf mm-wiki-ex/conf/
chmod -R 777 mm-wiki-ex
