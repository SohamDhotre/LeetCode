class Solution {
    public int findCircleNum(int[][] graph) {
        int count=0, m=graph.length, n=graph[0].length;
        int []vis=new int[m];
        for(int node=0;node<m;node++){
            if(vis[node]==0){
                count++;
                dfs(graph, node, vis);
            }
        }
        return count;
    }

    void dfs(int[][]graph, int node, int[]vis){
        vis[node]=1;
        for(int nextNode=0;nextNode<graph[0].length;nextNode++){
            if(graph[node][nextNode]==1 && vis[nextNode]==0){
                dfs(graph, nextNode, vis);
            }
        }
    }
}