class Solution {
    public boolean isBipartite(int[][] graph) {
        int n=graph.length;
        int[]color=new int[n];
        Arrays.fill(color, -1);
        for(int i=0;i<n;i++){
            if(color[i]==-1 && !bfsCheck(graph, i, color)) return false;
        }
        return true;
    }

    boolean bfsCheck(int [][]graph, int start, int []color){
        Queue<Integer>q=new ArrayDeque<>();
        q.offer(start);
        color[start]=0;
        while(!q.isEmpty()){
            int node=q.poll();
            for(int v:graph[node]){
                if(color[v]==-1){
                    color[v]=1-color[node];
                    q.offer(v);
                }
                else if(color[node]==color[v]) return false;
            }
        }
        return true;
    }
}