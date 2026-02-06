class Solution {
    int []parent;
    public int[] findRedundantConnection(int[][] edges) {
        parent=new int[edges.length+1];
        for(int node=1;node<parent.length;node++) 
            parent[node]=node;
        for(int []edge:edges)
            if(!union(edge[0], edge[1])) 
                return edge;

        return new int[0];
    }

    int findParent(int node){
        while(node!=parent[node])
            node=parent[node];

        return node;
    }

    boolean union(int node1, int node2) {
        int p1=findParent(node1);
        int p2=findParent(node2);
        if(p1==p2) return false;

        parent[p1]=p2;
        return true;
    }
}